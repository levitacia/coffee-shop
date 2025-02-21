package data

import "testing"

func TestCheckValidation(t *testing.T) {
	p := &Product{
		Name:  "Valera",
		Price: 1.24,
		SKU:   "iiii-iojj-kjeg",
	}

	err := p.Validate()

	if err != nil {
		t.Fatal(err)
	}
}
