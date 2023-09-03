package cake

import "testing"

type Service interface {
	Fruits() []string
	Veggies() []string
}

type LayerA struct{ Service }

func (l *LayerA) Fruits() []string {
	return []string{"Apple"}
}

func (l *LayerA) Veggies() []string {
	return []string{"Artichoke"}
}

type LayerB struct{ Service }

func (l *LayerB) Fruits() []string {
	return append(l.Service.Fruits(), "Banana")
}

func (l *LayerB) Veggies() []string {
	return append(l.Service.Veggies(), "Basil")
}

type LayerC struct{ Service }

func (l *LayerC) Veggies() []string {
	return append(l.Service.Veggies(), "Cilantro")
}

type LayerD struct{ Service }

func (l *LayerD) Fruits() []string {
	return append(l.Service.Fruits(), "Durian")
}

func (l *LayerD) Veggies() []string {
	return append(l.Service.Veggies(), "Dill")
}

type LayerE struct{ Service } // E for Empty!

func Test_Layers(t *testing.T) {
	testTable := map[string]struct {
		baseLayer       Service
		layers          []Service
		expectedFruits  []string
		expectedVeggies []string
	}{
		"Works as a wrapper even with no additional layers": {
			baseLayer:       &LayerA{},
			layers:          []Service{},
			expectedFruits:  []string{"Apple"},
			expectedVeggies: []string{"Artichoke"},
		},
		"Works as a wrapper one or more additional layers": {
			baseLayer: &LayerA{},
			layers: []Service{
				&LayerB{},
				&LayerC{},
				&LayerD{},
			},
			expectedFruits:  []string{"Apple", "Durian", "Banana"},
			expectedVeggies: []string{"Artichoke", "Dill", "Cilantro", "Basil"},
		},
		"Falls through layers with nil method implementations": {
			baseLayer: &LayerA{},
			layers: []Service{
				&LayerB{},
				&LayerC{},
				&LayerE{},
				&LayerD{},
			},
			expectedFruits:  []string{"Apple", "Durian", "Banana"},
			expectedVeggies: []string{"Artichoke", "Dill", "Cilantro", "Basil"},
		},
	}
	for name, testCase := range testTable {
		t.Run(name, func(t *testing.T) {
			svc, err := Layered[Service](testCase.baseLayer, testCase.layers...)
			if err != nil {
				t.Fatalf("failed to layer cake: %+v", err)
			}

			if len(svc.Fruits()) != len(testCase.expectedFruits) {
				t.Fatalf("expectedFruits %d words, got %d", len(testCase.expectedFruits), len(svc.Fruits()))
			}

			for i, fruit := range svc.Fruits() {
				if fruit != testCase.expectedFruits[i] {
					t.Fatalf("expectedFruits %s, got %s", testCase.expectedFruits[i], fruit)
				}
			}

			if len(svc.Veggies()) != len(testCase.expectedVeggies) {
				t.Fatalf("expectedVeggies %d words, got %d", len(testCase.expectedVeggies), len(svc.Veggies()))
			}

			for i, veg := range svc.Veggies() {
				if veg != testCase.expectedVeggies[i] {
					t.Fatalf("expectedVeggies %s, got %s", testCase.expectedVeggies[i], veg)
				}
			}
		})
	}
}
