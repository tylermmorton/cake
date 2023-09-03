package cake

import (
	"fmt"
	"reflect"
	"strings"
)

func Layered[T interface{}](base T, layers ...T) (T, error) {
	if len(layers) == 0 {
		return base, nil
	}

	var name = strings.Split(fmt.Sprintf("%T", new(T)), ".")[1]
	for i := 0; i < len(layers); i++ {
		val := reflect.ValueOf(layers[i]).Elem()
		field := val.FieldByName(name)
		if field.IsValid() && field.CanSet() {
			if i == len(layers)-1 {
				field.Set(reflect.ValueOf(base))
			} else {
				nextLayer := reflect.ValueOf(layers[i+1])
				field.Set(nextLayer)
			}
		}
	}
	return layers[0], nil
}
