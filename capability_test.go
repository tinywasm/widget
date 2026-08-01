package widget

import "testing"

// A filter control satisfies Filterable by declaring the setter alone — no
// embedding, no base type. This is the compile-time proof that the seam is
// nameable from outside without the control depending on anything here beyond
// the interface.
type fakeFilterControl struct{ sink func(term string) }

func (f *fakeFilterControl) OnFilterChange(fn func(term string)) { f.sink = fn }

func TestFilterable_SatisfiedByASetterAlone(t *testing.T) {
	var c Filterable = &fakeFilterControl{}

	got := ""
	c.OnFilterChange(func(term string) { got = term })

	c.(*fakeFilterControl).sink("backend")
	if got != "backend" {
		t.Errorf("sink did not receive the term: got %q, want %q", got, "backend")
	}
}

// The capability set is discovered by assertion, never by embedding a base
// type: a control that filters is not required to also select, dismiss or
// expand. Guards against someone later merging these into one fat interface.
func TestCapabilities_AreIndependent(t *testing.T) {
	var c any = &fakeFilterControl{}

	if _, ok := c.(Filterable); !ok {
		t.Error("expected the control to satisfy Filterable")
	}
	for name, ok := range map[string]bool{
		"Selectable":  func() bool { _, ok := c.(Selectable); return ok }(),
		"Dismissible": func() bool { _, ok := c.(Dismissible); return ok }(),
		"Expandable":  func() bool { _, ok := c.(Expandable); return ok }(),
	} {
		if ok {
			t.Errorf("a filter control must not be forced to satisfy %s", name)
		}
	}
}
