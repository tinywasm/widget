//go:build !wasm

package style_test

import (
	"strings"
	"testing"

	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
	"github.com/tinywasm/widget/style"
)

func TestCoverEmitsViewportFrame(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Root(style.Cover()).Stylesheet().String()

	// A definite height, not a floor: a min-height leaves the frame auto-sized,
	// so a Fill() descendant resolves to nothing and HideOverflow() never clips.
	if !strings.Contains(s, "height: 100dvh;") {
		t.Errorf("expected Cover to emit height: 100dvh, got:\n%s", s)
	}
	if strings.Contains(s, "min-height: 100dvh;") {
		t.Errorf("Cover must not fall back to min-height, got:\n%s", s)
	}
	if strings.Contains(s, "100vh") {
		t.Errorf("100vh (without d) must not appear in Cover output:\n%s", s)
	}
	if !strings.Contains(s, "@layer primitives") {
		t.Errorf("expected Cover to emit in @layer primitives:\n%s", s)
	}
}

func TestIconBoxEmitsSquareThatCannotShrink(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		Part("icon-sm", style.IconBox(style.IconSm)).
		Part("icon-md", style.IconBox(style.IconMd)).
		Part("icon-lg", style.IconBox(style.IconLg)).
		Stylesheet().String()

	for _, want := range []string{
		"width: 1em;", "height: 1em;",
		"width: 1.5em;", "height: 1.5em;",
		"width: 2.5em;", "height: 2.5em;",
		"flex-shrink: 0;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected IconBox to emit %q, got:\n%s", want, s)
		}
	}
}

func TestGrowClaimsWidthWithoutHeight(t *testing.T) {
	// Fill() also emits height: 100%, which inside a Row resolves against the
	// row and stretches the part into a full-height block. Grow() must not.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Part("label", style.Grow()).Stylesheet().String()

	if !strings.Contains(s, "flex-grow: 1;") {
		t.Errorf("expected Grow to emit flex-grow: 1, got:\n%s", s)
	}
	if !strings.Contains(s, "min-width: 0;") {
		t.Errorf("expected Grow to emit min-width: 0, got:\n%s", s)
	}
	if strings.Contains(s, "height: 100%;") {
		t.Errorf("Grow must not claim height, got:\n%s", s)
	}
}

func TestIconBoxStepsAreDistinct(t *testing.T) {
	w := testWidget{name: "shell", kind: widget.Region}
	seen := make(map[string]style.IconSize)
	for _, sz := range []style.IconSize{style.IconSm, style.IconMd, style.IconLg} {
		s := style.For(w).Part("icon", style.IconBox(sz)).Stylesheet().String()
		if prev, dup := seen[s]; dup {
			t.Errorf("IconSize %d and %d resolve to the same box", prev, sz)
		}
		seen[s] = sz
	}
}

func TestSidebarEndPutsRailLast(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	if !strings.Contains(s, "> :last-child") {
		t.Errorf("expected Sidebar(SideEnd) to emit > :last-child rail rule:\n%s", s)
	}
	if !strings.Contains(s, "flex-grow: 999;") {
		t.Errorf("expected Sidebar to emit flex-grow: 999 for content:\n%s", s)
	}
}

func TestSidebarStartMirrors(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideStart, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	if !strings.Contains(s, "> :first-child") {
		t.Errorf("expected Sidebar(SideStart) to emit > :first-child as rail:\n%s", s)
	}
}

func TestSidebarRailUsesToken(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Root(style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).Stylesheet().String()

	expected := css.RailNarrow.Var()
	if !strings.Contains(s, "--rail: "+expected) {
		t.Errorf("expected --rail value to be %q, got:\n%s", expected, s)
	}
}

func TestDrawerAnchorsToEdge(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Root(style.Drawer(style.SideEnd, style.TwoThirds), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, "position: fixed;") {
		t.Errorf("expected Drawer to emit position: fixed:\n%s", s)
	}
	if !strings.Contains(s, "inset-block: 0;") {
		t.Errorf("expected Drawer to emit inset-block: 0:\n%s", s)
	}
	if !strings.Contains(s, "inset-inline-end: 0;") {
		t.Errorf("expected Drawer(SideEnd) to emit inset-inline-end: 0:\n%s", s)
	}
	if !strings.Contains(s, "width: 66.66%;") {
		t.Errorf("expected Drawer(TwoThirds) to emit width: 66.66%%:\n%s", s)
	}
	if !strings.Contains(s, "z-index: var(--z-base") {
		t.Errorf("expected Drawer on Region to emit z-index matching layer:\n%s", s)
	}
}

func TestOnEmitsDeviceMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("menu", style.Row(style.Space1)).
		On(css.Mobile, "menu", style.Stack(style.Space2)).
		Stylesheet().String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected On(css.Mobile) to emit @media (max-width: 639.98px):\n%s", s)
	}
}

func TestOnlyOnHidesByDefault(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		OnlyOn(css.Mobile, "hamburger", style.Row(style.Space1)).
		Stylesheet().String()

	if !strings.Contains(s, ".w__hamburger {\n  display: none;\n}") {
		t.Errorf("expected OnlyOn part to emit display: none in widgets layer:\n%s", s)
	}
	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected OnlyOn to emit media block:\n%s", s)
	}
}

func TestOnlyOnPartPassesValidate(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		OnlyOn(css.Mobile, "hamburger", style.Row(style.Space1))

	if errs := sheet.Validate(); len(errs) > 0 {
		t.Errorf("expected OnlyOn part to pass Validate, got: %v", errs)
	}
}

func TestOnRevealedByStaysInsideMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure} // Disclosure allows Open
	s := style.For(w).
		Part("panel", style.RevealedBy(widget.Open)).
		On(css.Mobile, "panel", style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, "[data-open=\"true\"]") {
		t.Errorf("expected state rule for Open to be emitted:\n%s", s)
	}
}

func TestNoLayerStatementInsideMedia(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("menu", style.Row(style.Space1)).
		On(css.Mobile, "menu", style.Stack(style.Space2)).
		Stylesheet().String()

	idxMedia := strings.Index(s, "@media")
	if idxMedia >= 0 {
		after := s[idxMedia:]
		// Ensure @layer tokens, primitives, widgets, states does NOT appear after @media
		if strings.Contains(after, "@layer tokens,") {
			t.Errorf("found @layer statement inside @media block:\n%s", after)
		}
	}
}

func TestStateAttrsListsEveryRevealedBy(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Disclosure}
	sheet := style.For(w).
		Part("panel", style.RevealedBy(widget.Open)).
		Part("tab", style.RevealedBy(widget.Current)).
		On(css.Mobile, "panel", style.RevealedBy(widget.Open))

	attrs := sheet.StateAttrs()
	if len(attrs) != 2 {
		t.Errorf("expected 2 state attrs, got %d: %v", len(attrs), attrs)
	}
	found := make(map[string]bool)
	for _, a := range attrs {
		found[a.Key+"="+a.Value] = true
	}
	if !found["data-open=true"] {
		t.Error("expected data-open=true in StateAttrs")
	}
	if !found["data-current=true"] {
		t.Error("expected data-current=true in StateAttrs")
	}
}

func TestValidateDrawerWithoutRevealedBy(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		Part("panel", style.Drawer(style.SideEnd, style.TwoThirds))

	errs := sheet.Validate()
	if len(errs) == 0 {
		t.Fatal("expected Validate to reject Drawer without RevealedBy")
	}
	if !strings.Contains(errs[0].Error(), "Drawer() without RevealedBy()") {
		t.Errorf("unexpected error message: %v", errs[0])
	}
}

func TestEmissionDeterministic(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	sheet := style.For(w).
		Root(style.Cover()).
		Part("sidebar", style.Sidebar(style.SideEnd, style.RailNarrow, style.SpaceNone)).
		Part("drawer", style.Drawer(style.SideEnd, style.TwoThirds), style.RevealedBy(widget.Open)).
		On(css.Mobile, "drawer", style.Stack(style.Space2))

	css1 := sheet.Stylesheet().String()
	css2 := sheet.Stylesheet().String()
	if css1 != css2 {
		t.Errorf("Stylesheet emission is non-deterministic")
	}
}
