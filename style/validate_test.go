//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

// A part whose only job is a primitive emits into @layer primitives, never into
// @layer widgets. Validate must consult both layers before calling it empty.
func TestPrimitiveOnlyPartIsNotEmpty(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	cases := []struct {
		name string
		opt  style.Option
		decl string
	}{
		{"Fill", style.Fill(), "flex-grow: 1;"},
		{"Scroll", style.Scroll(), "overflow-y: auto;"},
		{"KeepSize", style.KeepSize(), "flex-shrink: 0;"},
		{"EdgeToEdge", style.EdgeToEdge(), "margin: 0;"},
		{"HideOverflow", style.HideOverflow(), "overflow: hidden;"},
		{"FillCentered", style.FillCentered(), "place-items: center;"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := style.For(w).Part("mount", c.opt)

			if errs := s.Validate(); len(errs) > 0 {
				t.Fatalf("%s alone must be a valid part, got: %v", c.name, errs)
			}

			out := s.Stylesheet().String()
			if !strings.Contains(out, c.decl) {
				t.Errorf("expected %q in emitted CSS, got:\n%s", c.decl, out)
			}
		})
	}
}

// The condition still catches the case it exists for: a part that declares nothing.
func TestPartWithoutOptionsEmitsNothing(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}

	errs := style.For(w).Part("mount").Validate()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one error, got: %v", errs)
	}
	if got := errs[0].Error(); got != `sheet w: part "mount" emits nothing` {
		t.Errorf("unexpected message: %s", got)
	}
}
