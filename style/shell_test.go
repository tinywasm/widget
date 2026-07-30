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

func TestStackGapSurvivesAFlowChild(t *testing.T) {
	// The separation must live on the container. A `> * + *` rule reading
	// var(--gap) resolves it against the CHILD, so a child that is itself a
	// flow container declares its own --gap and silently replaces the parent's
	// spacing — with SpaceNone it collapses it to nothing.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		Part("aside", style.Stack(style.Space1)).
		Part("aside-content", style.Stack(style.SpaceNone)).
		Stylesheet().String()

	if !strings.Contains(s, "gap: var(--gap);") {
		t.Errorf("expected Stack to put the gap on the container, got:\n%s", s)
	}
	if strings.Contains(s, "> * + *") {
		t.Errorf("Stack must not space its children with a child-resolved var(--gap), got:\n%s", s)
	}
}

func TestFlyoutHangsUnderItsAnchor(t *testing.T) {
	// A dropdown left in the flow pushes everything below it down, so the list
	// jumps under the pointer that opened it.
	w := testWidget{name: "shell", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Anchor()).
		Part("options", style.Flyout(style.SideEnd)).
		Stylesheet().String()

	for _, want := range []string{
		"position: relative;",
		"position: absolute;",
		"inset-block-start: 100%;",
		"inset-inline-end: 0;",
		"z-index: var(--z-dropdown,100);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Anchor/Flyout to emit %q, got:\n%s", want, s)
		}
	}
}

func TestDeviceRuleKeepsItsPrimitives(t *testing.T) {
	// The layout flags are grouped into shared selectors on the main path. A
	// device rule has nothing to group with, so they have to be emitted on its
	// own selector or an option passed to On()/OnlyOn() vanishes.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).
		OnlyOn(css.Mobile, "trigger", style.Row(style.Space1), style.PushEnd(), style.KeepSize()).
		Stylesheet().String()

	for _, want := range []string{"margin-inline-start: auto;", "flex-shrink: 0;"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected the device rule to keep %q, got:\n%s", want, s)
		}
	}
}

func TestMasterDetailRestsOnTheMaster(t *testing.T) {
	// The master has to be what shows on arrival, and it has to get there
	// without a scroll nudge: the component contract has no mount hook. RTL puts
	// the strip's start edge on the right, which is where scroll position 0
	// already is, and order:1 sends the master — second in the DOM — there.
	w := testWidget{name: "cv", kind: widget.Region}
	s := style.For(w).
		On(css.Mobile, "", style.MasterDetail(style.Most)).
		Stylesheet().String()

	for _, want := range []string{
		"direction: rtl;",
		"flex-wrap: nowrap;",
		"flex: 0 0 auto;",
		"scroll-snap-type: x mandatory;",
		".cv > :nth-child(2)",
		"order: 1;",
		"scroll-snap-align: start;",
		".cv > :nth-child(1)",
		"order: 2;",
		"scroll-snap-align: end;",
		"flex: 0 0 90%;",
		"direction: ltr;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected MasterDetail to emit %q, got:\n%s", want, s)
		}
	}
	if !strings.Contains(s, "@media") {
		t.Errorf("expected the rule to stay inside its device query, got:\n%s", s)
	}
}

func TestPushEndMovesFreeSpaceInFront(t *testing.T) {
	// Once flex-wrap drops an item onto a line of its own, nothing else on that
	// line can push it: the free space has to go in front of it.
	w := testWidget{name: "shell", kind: widget.Region}
	s := style.For(w).Part("badge", style.PushEnd()).Stylesheet().String()

	if !strings.Contains(s, "margin-inline-start: auto;") {
		t.Errorf("expected PushEnd to emit margin-inline-start: auto, got:\n%s", s)
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
