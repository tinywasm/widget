//go:build !wasm

package style

import (
	"strings"
	"testing"

	"github.com/tinywasm/widget"
)

func TestOverlay(t *testing.T) {
	// 1. Of("w").Root(Overlay(Viewport)) emite position: fixed y inset: 0 y z-index: 100
	s1 := Of(widget.Name("w")).Root(Overlay(Viewport)).Stylesheet().String()
	if !strings.Contains(s1, "position: fixed;") {
		t.Errorf("expected s1 to contain 'position: fixed;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "inset: 0;") {
		t.Errorf("expected s1 to contain 'inset: 0;', got:\n%s", s1)
	}
	if !strings.Contains(s1, "z-index: 100;") {
		t.Errorf("expected s1 to contain 'z-index: 100;', got:\n%s", s1)
	}

	// 2. Of("w").Root(Overlay(Parent)) emite position: absolute (no fixed)
	s2 := Of(widget.Name("w")).Root(Overlay(Parent)).Stylesheet().String()
	if !strings.Contains(s2, "position: absolute;") {
		t.Errorf("expected s2 to contain 'position: absolute;', got:\n%s", s2)
	}
	if strings.Contains(s2, "position: fixed;") {
		t.Errorf("expected s2 NOT to contain 'position: fixed;', got:\n%s", s2)
	}

	// 3. Of("w").Part("p", Above()) emite z-index: 101
	s3 := Of(widget.Name("w")).Part(widget.Part("p"), Above()).Stylesheet().String()
	if !strings.Contains(s3, "z-index: 101;") {
		t.Errorf("expected s3 to contain 'z-index: 101;', got:\n%s", s3)
	}

	// 4. Of("w").Root(Scrim()) emite un background-color con color-mix( y var(--color-surface), y no contiene ningún literal hexadecimal
	s4 := Of(widget.Name("w")).Root(Scrim()).Stylesheet().String()
	if !strings.Contains(s4, "background-color:") {
		t.Errorf("expected s4 to contain 'background-color:', got:\n%s", s4)
	}
	if !strings.Contains(s4, "color-mix(") {
		t.Errorf("expected s4 to contain 'color-mix(', got:\n%s", s4)
	}
	if !strings.Contains(s4, "var(--color-surface)") {
		t.Errorf("expected s4 to contain 'var(--color-surface)', got:\n%s", s4)
	}
	if strings.Contains(s4, "#") {
		t.Errorf("expected s4 NOT to contain '#', got:\n%s", s4)
	}

	// 5. El caso que motiva el plan: una hoja con
	// .Part(p, Overlay(Viewport), Hidden()).When(widget.Open, p, Shown())
	// emite display: none dentro de @layer widgets y display: block dentro de @layer states, en ese orden, y
	// el selector de estado es .w__p[data-open="true"]
	s5 := Of(widget.Name("w")).
		Part(widget.Part("p"), Overlay(Viewport), Hidden()).
		When(widget.Open, widget.Part("p"), Shown()).
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

	// 6. La hoja emitida no contiene :has(
	if strings.Contains(s5, ":has(") {
		t.Error("expected sheet NOT to contain ':has('")
	}
}
