package cake

import (
	"fmt"
	"reflect"
	"strings"
)

func getLayerValue(layer any) (reflect.Value, bool) {
	if layer == nil {
		return reflect.Value{}, false
	}

	val := reflect.ValueOf(layer)
	if val.IsZero() {
		return val, false
	} else if val.IsNil() {
		return val, false
	} else if val.Kind() != reflect.Ptr {
		return val, false
	}

	return val, true
}

// If returns the layer if cond is true, otherwise it returns a zero value of the layer's type.
// This is useful for skipping entire layers based on a condition.
func If[T interface{}](cond bool, layer T) T {
	if cond {
		return layer
	} else {
		return *new(T)
	}
}

// IfCallback returns the result of the layer function if cond is true, otherwise it returns a
// zero value of the layer's type. This is useful for skipping entire layers based on a condition
// when the layer is expensive to construct.
func IfCallback[T interface{}](cond bool, layer func() T) T {
	if cond {
		return layer()
	} else {
		return *new(T)
	}
}

// Layered takes base layer T and a list of additional layers and constructs a single T value
// that is a wrapper around the base layer. The layers are applied in order, with the last layer
// being the outermost layer. This is useful for wrapping a base layer with additional functionality
// without having to modify the base layer.
func Layered[T interface{}](base T, layers ...T) (T, error) {
	if len(layers) == 0 {
		return base, nil
	}

	var entryLayer = -1
	// get the name of T, which is the interface that all layers implement
	var interfaceName = strings.Split(fmt.Sprintf("%T", new(T)), ".")[1]

	// iterate through all provided layers.
	for i := 0; i < len(layers); i++ {
		// layers should be a pointer to a struct that implements T
		curLayerValue, ok := getLayerValue(layers[i])
		if !ok {
			continue
		}

		if entryLayer == -1 {
			entryLayer = i
		}

		curLayerValue = curLayerValue.Elem()

		// get a reference to the value of the embedded field that
		// implements the interface that T represents
		targetField := curLayerValue.FieldByName(interfaceName)
		if !targetField.IsValid() || !targetField.CanSet() {
			return *new(T), fmt.Errorf("field %s in layer '%T' cannot be set", targetField.String(), layers[i])
		}

		// if this is the last provided layer, set the embedded field to the base layer
		if i == len(layers)-1 {
			targetField.Set(reflect.ValueOf(base))
			break
		}

		// now seek the next valid layer
		var didSet bool
		for j := i + 1; j < len(layers); j++ {
			nextLayerValue, ok := getLayerValue(layers[j])
			if !ok {
				continue
			}

			didSet = true
			targetField.Set(nextLayerValue)

			// i will be incremented just after this
			// so set it to the previous layer
			i = j - 1

			break
		}

		// if after iterating through all the layers we did not set the value of the embedded field,
		// set it to the base layer's value and break the loop.
		if !didSet {
			targetField.Set(reflect.ValueOf(base))
			break
		}
	}

	return layers[entryLayer], nil
}
