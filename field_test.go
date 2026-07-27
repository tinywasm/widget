package widget

import (
	"testing"

	"github.com/tinywasm/fmt"
)

func TestFieldAnatomy(t *testing.T) {
	// 1. NameField.Root().String() is exactly "tw-field"
	if got, want := NameField.Root().String(), "tw-field"; got != want {
		t.Errorf("NameField.Root().String() = %q; want %q", got, want)
	}

	// 2. NameField.Class(PartLabel).String() is exactly "tw-field__label"
	if got, want := NameField.Class(PartLabel).String(), "tw-field__label"; got != want {
		t.Errorf("NameField.Class(PartLabel).String() = %q; want %q", got, want)
	}

	// 3. Same for PartInput, PartError, and PartRadioGroup
	if got, want := NameField.Class(PartInput).String(), "tw-field__input"; got != want {
		t.Errorf("NameField.Class(PartInput).String() = %q; want %q", got, want)
	}

	if got, want := NameField.Class(PartError).String(), "tw-field__error"; got != want {
		t.Errorf("NameField.Class(PartError).String() = %q; want %q", got, want)
	}

	if got, want := NameField.Class(PartRadioGroup).String(), "tw-field__radio-group"; got != want {
		t.Errorf("NameField.Class(PartRadioGroup).String() = %q; want %q", got, want)
	}

	// 4. NameField.Class(PartLabel).AsAttr() returns fmt.KeyValue{Key: "class", Value: "tw-field__label"}
	expectedAttr := fmt.KeyValue{Key: "class", Value: "tw-field__label"}
	if got := NameField.Class(PartLabel).AsAttr(); got != expectedAttr {
		t.Errorf("NameField.Class(PartLabel).AsAttr() = %+v; want %+v", got, expectedAttr)
	}
}
