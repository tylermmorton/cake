package cake

import "testing"

type Service interface {
	Words() []string
}

type LayerA struct{ Service }

func (l *LayerA) Words() []string {
	return []string{"Artichoke"}
}

type LayerB struct{ Service }

func (l *LayerB) Words() []string {
	return append(l.Service.Words(), "Basil")
}

type LayerC struct{ Service }

func (l *LayerC) Words() []string {
	return append(l.Service.Words(), "Cilantro")
}

type LayerD struct{ Service }

func (l *LayerD) Words() []string {
	return append(l.Service.Words(), "Dill")
}

type LayerE struct{ Service } // E for Empty!

func Test_Layers(t *testing.T) {
	testTable := map[string]struct {
		baseLayer Service
		layers    []Service
		expected  []string
	}{
		"Works as a wrapper even with no additional layers": {
			baseLayer: &LayerA{},
			layers:    []Service{},
			expected:  []string{"Artichoke"},
		},
		"Works as a wrapper one or more additional layers": {
			baseLayer: &LayerA{},
			layers: []Service{
				&LayerB{},
				&LayerC{},
				&LayerD{},
			},
			expected: []string{"Artichoke", "Dill", "Cilantro", "Basil"},
		},
		"Falls through empty layers": {
			baseLayer: &LayerA{},
			layers: []Service{
				&LayerB{},
				&LayerC{},
				&LayerE{},
				&LayerD{},
			},
			expected: []string{"Artichoke", "Dill", "Cilantro", "Basil"},
		},
	}
	for name, testCase := range testTable {
		t.Run(name, func(t *testing.T) {
			svc, err := Layered[Service](testCase.baseLayer, testCase.layers...)
			if err != nil {
				t.Fatalf("failed to layer cake: %+v", err)
			}

			if len(svc.Words()) != len(testCase.expected) {
				t.Fatalf("expected %d words, got %d", len(testCase.expected), len(svc.Words()))
			}

			for i, word := range svc.Words() {
				if word != testCase.expected[i] {
					t.Fatalf("expected %s, got %s", testCase.expected[i], word)
				}
			}
		})
	}
}
