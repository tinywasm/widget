//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

type testWidget struct {
	name widget.Name
	kind widget.Kind
}

func (t testWidget) WidgetName() widget.Name { return t.name }
func (t testWidget) WidgetKind() widget.Kind { return t.kind }

func TestBackdrop(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Dialog}

	// 1. style.For(w).Root(Backdrop(Viewport)) emits position: fixed and inset: 0 and z-index: var(--z-modal,300);
	s1 := style.For(w).Root(style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(s1, "position: fixed;") {
		t.Errorf("expected s1 to contain 'position: fixed;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "inset: 0;") {
		t.Errorf("expected s1 to contain 'inset: 0;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "z-index: var(--z-modal,300);") {
		t.Errorf("expected s1 to contain 'z-index: var(--z-modal,300);', got:\n%s", s1)
	}

	// 2. style.For(w).Root(Backdrop(Parent)) emits position: absolute (no fixed)
	s2 := style.For(w).Root(style.Backdrop(style.Parent)).Stylesheet().String()
	if !strings.Contains(s2, "position: absolute;") {
		t.Errorf("expected s2 to contain 'position: absolute;', got:\n%s", s2)
	}
	if strings.Contains(s2, "position: fixed;") {
		t.Errorf("expected s2 NOT to contain 'position: fixed;', got:\n%s", s2)
	}

	// 3. style.For(w).Root(Veil(), Backdrop(Viewport)) emits a background-color with color-mix
	s4 := style.For(w).Root(style.Veil(), style.Backdrop(style.Viewport)).Stylesheet().String()
	if !strings.Contains(s4, "background-color:") {
		t.Errorf("expected s4 to contain 'background-color:', got:\n%s", s4)
	}
	if !strings.Contains(s4, "color-mix(") {
		t.Errorf("expected s4 to contain 'color-mix(', got:\n%s", s4)
	}
	// Not var(--color-surface: the color-mix() argument is
	// css.ColorSurface.NestedEnhanced(), a literal light-dark(...) with no
	// var() anywhere — a var() here, even to an always-safe property, would
	// defer this whole declaration's validity to computed-value time and
	// break the legacy-Safari fallback (see css.Token.NestedEnhanced).
	if !strings.Contains(s4, "light-dark(") {
		t.Errorf("expected s4 to contain 'light-dark(', got:\n%s", s4)
	}

	// 4. RevealedBy case
	s5 := style.For(w).
		Part(widget.Part("p"), style.Backdrop(style.Viewport), style.RevealedBy(widget.Open)).
		Stylesheet().
		String()

	idxWidgets := strings.Index(s5, "@layer widgets {")
	idxStates := strings.Index(s5, "@layer states {")
	if idxWidgets == -1 {
		t.Error("expected to find @layer widgets")
	}
	if idxStates == -1 {
		t.Error("expected to find @layer states")
	}
	if idxWidgets >= idxStates {
		t.Errorf("expected @layer widgets to come before @layer states in:\n%s", s5)
	}

	widgetsPart := s5[idxWidgets:idxStates]
	statesPart := s5[idxStates:]

	if !strings.Contains(widgetsPart, ".w__p {") || !strings.Contains(widgetsPart, "display: none;") {
		t.Errorf("expected .w__p with display: none; in @layer widgets part, got:\n%s", widgetsPart)
	}

	if !strings.Contains(statesPart, ".w__p[data-open=\"true\"] {") || !strings.Contains(statesPart, "display: block;") {
		t.Errorf("expected .w__p[data-open=\"true\"] with display: block; in @layer states part, got:\n%s", statesPart)
	}
}
