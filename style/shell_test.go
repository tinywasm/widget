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
		".cv > *",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected MasterDetail to emit %q, got:\n%s", want, s)
		}
	}
	if !strings.Contains(s, "@media") {
		t.Errorf("expected the rule to stay inside its device query, got:\n%s", s)
	}
}

func TestDockedPinsToTheAnchorCorner(t *testing.T) {
	w := testWidget{name: "cv", kind: widget.Disclosure}
	s := style.For(w).
		Part("aside", style.Anchor()).
		Part("action", style.Docked(style.Parent, style.EdgeBottom, style.SideEnd, style.Space4), style.CenterContent()).
		Stylesheet().String()

	for _, want := range []string{
		"position: relative;",
		"position: absolute;",
		"inset-block-end: var(--space-4,1rem);",
		"inset-inline-end: var(--space-4,1rem);",
		"justify-content: center;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Docked/CenterContent to emit %q, got:\n%s", want, s)
		}
	}

	// A Parent dock stays out of the overlay layer: siblings doing the same
	// would tie there, and the last in the DOM would cover the panel the first
	// one opened. It still declares a LOCAL level (z-index: 1) — leaving it
	// `auto` let Safari paint an unpositioned sibling over the pinned control
	// — and that is the whole declaration it is allowed: any other level is a
	// claim on a layer it does not own.
	if !strings.Contains(s, "z-index: 1;") {
		t.Errorf("Docked(Parent) must declare its local stacking level, got:\n%s", s)
	}
	if strings.Contains(s, "z-index: var(--z-") {
		t.Errorf("Docked(Parent) must not claim an overlay stacking level, got:\n%s", s)
	}

	// A corner pin owns all four insets: the unpinned pair has to be reset, or
	// overriding another positioning option leaves the box over-constrained and
	// it collapses instead of moving.
	for _, want := range []string{"inset-block-start: auto;", "inset-inline-start: auto;"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Docked to reset the unpinned insets with %q, got:\n%s", want, s)
		}
	}

	// A Viewport dock is a real overlay and does claim it.
	v := style.For(w).Part("fab", style.Docked(style.Viewport, style.EdgeBottom, style.SideEnd, style.Space4)).Stylesheet().String()
	if !strings.Contains(v, "z-index: var(--z-dropdown,100);") {
		t.Errorf("expected Docked(Viewport) to claim the widget layer, got:\n%s", v)
	}
}

func TestOnEdgeStraddlesTheLine(t *testing.T) {
	// A fieldset legend rides the border it labels. Half of the chip has to sit
	// outside the box and half inside; the straddle is exact because the chip's
	// height is the shared --chip-height token, applied as half a chip-height
	// of negative margin — not a transform, which is invisible to scrollHeight
	// by spec (no ancestor could reserve the space the chip really occupies)
	// and which created an implicit stacking context.
	w := testWidget{name: "tw-field", kind: widget.Form}
	s := style.For(w).
		Root(style.Anchor()).
		Part("label", style.OnEdge(style.EdgeTop, style.SideStart, style.Space2, style.Space4)).
		Part("badge", style.OnEdge(style.EdgeBottom, style.SideEnd, style.SpaceNone, style.Space3)).
		Stylesheet().String()

	for _, want := range []string{
		"inset-block-start: var(--space-2,0.5rem);",
		"inset-inline-start: var(--space-4,1rem);",
		"margin-block-start: calc(-0.5 * " + css.ChipHeight.Var() + ");",
		"inset-block-end: 0;",
		"inset-inline-end: var(--space-3,0.75rem);",
		"margin-block-end: calc(-0.5 * " + css.ChipHeight.Var() + ");",
		"margin: 0;",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected OnEdge to emit %q, got:\n%s", want, s)
		}
	}
	if strings.Contains(s, "transform") {
		t.Errorf("OnEdge must not emit transform, got:\n%s", s)
	}

	// A chip is content, so it must not reach the OVERLAY layer: level with the
	// real overlays it wins on DOM order, which puts it over a dropdown. That —
	// not "no z-index at all" — is the invariant. It does order itself against
	// its own siblings with a local z-index: leaving it `auto` let Safari paint
	// a sibling form control over the legend labelling it.
	if !strings.Contains(s, "z-index: 1;") {
		t.Errorf("expected OnEdge to order itself locally with z-index: 1, got:\n%s", s)
	}
	if strings.Contains(s, "z-index: var(--z-") {
		t.Errorf("OnEdge must not claim an overlay stacking level, got:\n%s", s)
	}
}

func TestGlyphTintsWithoutFilling(t *testing.T) {
	// A nav item that is merely available shows a coloured icon; only the
	// selected one gets the filled surface. Icons follow currentColor, so the
	// tint has to reach fill as well as color.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).Part("nav-link", style.Glyph(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, "fill: currentColor;") {
		t.Errorf("expected Glyph to reach the icon through fill, got:\n%s", s)
	}
	if strings.Contains(s, "background-color") {
		t.Errorf("Glyph must not paint a background, got:\n%s", s)
	}
}

func TestControlBoxSharesOneHeight(t *testing.T) {
	// A list row and a form field have to be measured against the same token or
	// they drift apart the moment either one's padding changes.
	w := testWidget{name: "x", kind: widget.Region}
	s := style.For(w).
		Part("row", style.ControlBox()).
		Part("field", style.ControlBox()).
		Stylesheet().String()

	if !strings.Contains(s, "min-height: "+css.ControlHeight.Var()+";") {
		t.Errorf("expected ControlBox to emit the shared height token, got:\n%s", s)
	}
}

func TestCueWithinReachesADescendant(t *testing.T) {
	// The one descendant selector in the package. A rail that reveals its own
	// labels on hover has no other expression: Cue() only ever emits
	// `.n__part:hover`, and the label is a different element from the trigger.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Stack(style.Space1)).
		Part("link-text", style.FontSize(style.TextBase)).
		CueWithin(widget.Hover, "menu", "link-text", style.Row(style.SpaceNone)).
		Stylesheet().String()

	if !strings.Contains(s, ".pd__menu:hover .pd__link-text") {
		t.Errorf("expected the descendant selector, got:\n%s", s)
	}
}

func TestCueWithinHoverScopesToFinePointer(t *testing.T) {
	// CueWithinHover is the same descendant selector as CueWithin, gated on
	// the fine-pointer capability: a touch tap fires :hover but never
	// (hover: hover), so a hover reveal scoped here cannot misfire on a phone.
	w := testWidget{name: "pd", kind: widget.Menu}
	s := style.For(w).
		Part("menu", style.Stack(style.Space1)).
		Part("drawer-panel", style.Raise(style.Floating)).
		CueWithinHover(widget.Hover, "menu", "drawer-panel", style.Row(style.SpaceNone)).
		Stylesheet().String()

	mediaIdx := strings.Index(s, "@media (hover: hover)")
	if mediaIdx < 0 {
		t.Fatalf("expected the fine-pointer gate, got:\n%s", s)
	}
	block := s[mediaIdx:]
	if !strings.Contains(block, ".pd__menu:hover .pd__drawer-panel") {
		t.Errorf("expected the descendant selector inside the hover media query, got:\n%s", s)
	}
	if !strings.Contains(block, "@layer states") {
		t.Errorf("expected the rule in @layer states (so it outranks @layer widgets device rules), got:\n%s", block)
	}
	plain := s[:mediaIdx]
	if strings.Contains(plain, ".pd__menu:hover .pd__drawer-panel") {
		t.Errorf("the selector must not escape into the plain states layer, got:\n%s", plain)
	}
}

func TestValidateCueWithinHoverContainerUndeclared(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Menu}
	sheet := style.For(w).
		Part("link-text", style.FontSize(style.TextBase)).
		CueWithinHover(widget.Hover, "menu", "link-text", style.Row(style.SpaceNone))

	errs := sheet.Validate()
	if len(errs) == 0 {
		t.Fatal("expected Validate to reject an undeclared CueWithinHover container")
	}
	if !strings.Contains(errs[0].Error(), `CueWithinHover container "menu" is not a declared part`) {
		t.Errorf("unexpected error message: %v", errs[0])
	}
}

func TestSlideDeckParksPagesOffCanvas(t *testing.T) {
	// SlideDeck is not a scroller: each child is an absolute layer parked at the
	// inline-start edge, and the current one is the only one in the box.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	for _, want := range []string{
		".w__x > *",
		"position: absolute;",
		"inset: 0;",
		"transform: translateX(-100%);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected SlideDeck pages to emit %q, got:\n%s", want, s)
		}
	}

	// Exactly three rules for the part: container, pages, current page.
	if n := strings.Count(s, ".w__x "); n != 3 {
		t.Errorf("expected exactly 3 rules for a SlideDeck part, got %d:\n%s", n, s)
	}
}

func TestSlideDeckRevealsTheCurrentPage(t *testing.T) {
	// The state is widget.Current, derived from its Attr(), so markup and CSS
	// match by construction — the same principle that sustains RevealedBy.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	if !strings.Contains(s, `.w__x > *[data-current="true"]`) {
		t.Errorf("expected the current-page selector, got:\n%s", s)
	}
	if !strings.Contains(s, "transform: translateX(0);") {
		t.Errorf("expected the current page to sit in the box, got:\n%s", s)
	}
}

func TestSlideDeckIsNotAScroller(t *testing.T) {
	// The whole point: no scroller here, or the gesture inside a module chains
	// with the stage's own snap strip and changes section alone.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.SlideDeck(style.MotionBase)).Stylesheet().String()

	if strings.Contains(s, "scroll-snap-type") {
		t.Errorf("SlideDeck must not emit scroll-snap-type, got:\n%s", s)
	}
	if strings.Contains(s, "overflow-x") {
		t.Errorf("SlideDeck must not emit overflow-x, got:\n%s", s)
	}
}

func TestAutoRotateRunsOnZeroValueReceiver(t *testing.T) {
	// The package-wide contract: RenderCSS runs on &T{}, so AutoRotate cannot
	// take a count and must never panic or need instance data to build its
	// stylesheet.
	w := testWidget{name: "w", kind: widget.Region}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AutoRotate panicked on a zero-value receiver: %v", r)
		}
	}()
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()
	if !strings.Contains(s, "@keyframes tw-auto-rotate") {
		t.Errorf("expected the shared keyframes rule, got:\n%s", s)
	}
}

func TestAutoRotateStagersDelayByPosition(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()

	for _, want := range []string{
		".w__x > *",
		"position: absolute;",
		"inset: 0;",
		"opacity: 0;",
		"animation: tw-auto-rotate",
		".w__x > :first-child {\n  opacity: 1;\n}",
		".w__x > :nth-child(2) {\n  animation-delay: 5s;\n}",
		".w__x > :nth-child(6) {\n  animation-delay: 25s;\n}",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected AutoRotate to emit %q, got:\n%s", want, s)
		}
	}

	// No 7th slot: AutoRotateLayers caps the stagger.
	if strings.Contains(s, ":nth-child(7)") {
		t.Errorf("expected no :nth-child(7) rule beyond AutoRotateLayers, got:\n%s", s)
	}
}

func TestAutoRotateRespectsReducedMotion(t *testing.T) {
	// Disabling the animation must leave :first-child visible and every other
	// layer hidden — the "first image, frozen" fallback — with no JS and no
	// extra markup from the caller.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.AutoRotate()).Stylesheet().String()

	if !strings.Contains(s, "@media (prefers-reduced-motion: reduce)") {
		t.Fatalf("expected a prefers-reduced-motion block, got:\n%s", s)
	}
	if !strings.Contains(s, ".w__x > * {\n  animation: none;\n}") {
		t.Errorf("expected AutoRotate layers to disable their animation under reduced motion, got:\n%s", s)
	}
}

func TestStateSurfaceRepaintsBorderAsRing(t *testing.T) {
	// A state is painted OVER the base box: a border here grows the element
	// exactly when the pointer is on it, and the hover is what made it grow.
	// The border is therefore repainted as a shadow ring — 0 0 0 1px, the
	// border width a hair's breadth outside the box — never as a border, and
	// never as an outline (see the ring comment in emit_decls.go for why).
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("x", style.As(style.Page)).
		Cue(widget.Hover, "x", style.As(style.Inset)).
		Stylesheet().String()

	idx := strings.Index(s, ".w__x:hover")
	if idx < 0 {
		t.Fatalf("expected a :hover rule, got:\n%s", s)
	}
	hoverRule := s[idx : idx+strings.Index(s[idx:], "}")]

	if !strings.Contains(hoverRule, "box-shadow: 0 0 0 1px") {
		t.Errorf("expected the state rule to repaint the border as a ring, got:\n%s", hoverRule)
	}
	if strings.Contains(hoverRule, "border:") {
		t.Errorf("a state rule must not emit border:, got:\n%s", hoverRule)
	}
	// "outline:" with the colon, not the bare word: the ring's colour is
	// var(--color-outline), and matching the token NAME would flag the very
	// declaration this test wants.
	if strings.Contains(hoverRule, "outline:") {
		t.Errorf("a state rule must not emit outline, got:\n%s", hoverRule)
	}
}

func TestStateRingComposesWithElevation(t *testing.T) {
	// Raise() and a state border both paint through box-shadow. When a state
	// rule raises AND repaints its border, the two must compose into ONE
	// declaration — ring first, elevation after — or the later declaration
	// would silently stomp the earlier one and the raised border would come
	// out bare.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).
		Part("x", style.As(style.Page), style.Raise(style.Raised)).
		Cue(widget.Hover, "x", style.As(style.Inset), style.Raise(style.Raised)).
		Stylesheet().String()

	idx := strings.Index(s, ".w__x:hover")
	if idx < 0 {
		t.Fatalf("expected a :hover rule, got:\n%s", s)
	}
	hoverRule := s[idx : idx+strings.Index(s[idx:], "}")]

	// The ring is the token's Var() form when it composes with an elevation
	// (see boxShadowDecls): the light-dark() half would defer the whole
	// declaration to computed-value time, so the static/enhanced pair is
	// structurally impossible here.
	want := "box-shadow: 0 0 0 1px " + css.ColorOutline.Var() + ", " + css.ShadowSm.Var() + ";"
	if !strings.Contains(hoverRule, want) {
		t.Errorf("expected the raised state rule to compose ring then elevation in one declaration:\nwant %q\ngot:\n%s", want, hoverRule)
	}
	if strings.Contains(hoverRule, "box-shadow: "+css.ShadowSm.Var()+";") {
		t.Errorf("a raised state rule must not repeat a bare elevation declaration, got:\n%s", hoverRule)
	}
}

func TestFloatingChromeReservesScrollEndSpace(t *testing.T) {
	// A floating chrome strip (the host: a FAB, a miniplayer) declares the
	// band of its edge it occupies as an inherited custom property; every
	// Scroll() region — in this widget or a descendant one — reserves it
	// through var(--floating-bottom, 0px). No chrome means no declaration and
	// no reservation.
	w := testWidget{name: "fab", kind: widget.Region}
	s := style.For(w).
		Part("host", style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4)).
		Part("list", style.Scroll()).
		Stylesheet().String()

	want := "--floating-bottom: calc(2.5em + 2 * " + css.Space4.Var() + ");"
	if !strings.Contains(s, want) {
		t.Errorf("expected the host to declare its strip with %q, got:\n%s", want, s)
	}
	for _, want := range []string{
		"padding-block-start: var(--floating-top, 0px);",
		"padding-block-end: var(--floating-bottom, 0px);",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Scroll() to reserve the floating strip with %q, got:\n%s", want, s)
		}
	}

	// EdgeTop is the mirror: the strip declaration switches property, the
	// Scroll() padding stays the same.
	top := style.For(w).Part("host", style.FloatingChrome(style.EdgeTop, style.IconLg, style.Space4)).Stylesheet().String()
	if !strings.Contains(top, "--floating-top: calc(2.5em + 2 * "+css.Space4.Var()+");") {
		t.Errorf("expected FloatingChrome(EdgeTop) to declare --floating-top, got:\n%s", top)
	}

	// And with no chrome at all, Scroll() pads by the 0px default — the var
	// references stay, but no sheet declares the property.
	bare := style.For(w).Part("list", style.Scroll()).Stylesheet().String()
	if strings.Contains(bare, "--floating-top:") || strings.Contains(bare, "--floating-bottom:") {
		t.Errorf("no FloatingChrome means no floating declaration, got:\n%s", bare)
	}
}

func TestScrollGutterAddsToTheFloatingReservationWithoutReplacingIt(t *testing.T) {
	// ScrollGutter exists so a consumer (tinywasm/components/listgap) can give
	// a Scroll() region its own ambient top/bottom gutter WITHOUT the plain
	// widgets-layer PadEdge()/Pad() trap: a later-layer declaration replaces
	// an earlier one outright (CSS layers do not add), which would silently
	// erase whatever a FloatingChrome ancestor reserved. Folding both into one
	// calc is what keeps them both live at once.
	w := testWidget{name: "fab", kind: widget.Region}

	// With an active FloatingChrome reservation, the gutter must be ADDED to
	// it, not replace it.
	s := style.For(w).
		Part("host", style.FloatingChrome(style.EdgeBottom, style.IconLg, style.Space4)).
		Part("list", style.Scroll(), style.ScrollGutter(style.Space1)).
		Stylesheet().String()

	want := "padding-block-end: calc(var(--floating-bottom, 0px) + " + css.Space1.Var() + ");"
	if !strings.Contains(s, want) {
		t.Errorf("expected the gutter to add to the floating reservation with %q, got:\n%s", want, s)
	}
	// The top edge carries the SAME gutter even though only EdgeBottom has an
	// active FloatingChrome — the gutter is ambient on both edges by design.
	wantTop := "padding-block-start: calc(var(--floating-top, 0px) + " + css.Space1.Var() + ");"
	if !strings.Contains(s, wantTop) {
		t.Errorf("expected the gutter on the top edge too with %q, got:\n%s", wantTop, s)
	}

	// A Scroll() region with no gutter must keep emitting the exact old decl —
	// byte-identical — so every existing consumer is unaffected.
	bare := style.For(w).Part("list", style.Scroll()).Stylesheet().String()
	if !strings.Contains(bare, "padding-block-end: var(--floating-bottom, 0px);") {
		t.Errorf("a plain Scroll() must keep the old decl untouched, got:\n%s", bare)
	}
	if strings.Contains(bare, "calc(var(--floating-bottom") {
		t.Errorf("a plain Scroll() must never emit the calc form, got:\n%s", bare)
	}
}

func TestBaseSurfaceKeepsBorder(t *testing.T) {
	// Only STATE rules switch to outline; the base box still owns its border.
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("x", style.As(style.Inset)).Stylesheet().String()

	if !strings.Contains(s, "border: 1px solid") {
		t.Errorf("expected the base rule to emit border:, got:\n%s", s)
	}
	if strings.Contains(s, "outline:") {
		t.Errorf("a base rule must not emit outline, got:\n%s", s)
	}
}

func TestHideSwitchesAPartOffForOneDevice(t *testing.T) {
	// OnlyOn hides by default and reveals per device. Hide() is the other
	// direction: keep the base styling, switch the part off on one device.
	w := testWidget{name: "pd", kind: widget.Region}
	s := style.For(w).
		Part("header", style.Row(style.Space2), style.As(style.Panel)).
		On(css.Mobile, "header", style.Hide()).
		Stylesheet().String()

	if !strings.Contains(s, "@media (max-width: 639.98px)") {
		t.Errorf("expected the rule to be device-scoped, got:\n%s", s)
	}
	if !strings.Contains(s, "display: none;") {
		t.Errorf("expected Hide to emit display:none, got:\n%s", s)
	}
	if !strings.Contains(s, "background-color") {
		t.Errorf("the base styling must survive, got:\n%s", s)
	}
}

func TestChipBoxSharesOneWidth(t *testing.T) {
	w := testWidget{name: "x", kind: widget.Region}
	s := style.For(w).Part("badge", style.ChipBox()).Stylesheet().String()

	if !strings.Contains(s, "width: "+css.ChipWidth.Var()+";") {
		t.Errorf("expected ChipBox to emit the shared width token, got:\n%s", s)
	}
	if !strings.Contains(s, "overflow: hidden;") {
		t.Errorf("expected ChipBox to clip, got:\n%s", s)
	}
}

func TestRevealedSplitStaysFlex(t *testing.T) {
	// Split emits display:flex in @layer primitives; a state rule saying "grid"
	// would win from @layer states and strand every flex-basis under it.
	w := testWidget{name: "w", kind: widget.Disclosure}
	s := style.For(w).
		Part("panel", style.Split(style.SplitTwoThirds, style.Space2), style.RevealedBy(widget.Open)).
		Stylesheet().String()

	if !strings.Contains(s, ".w__panel[data-open=\"true\"] {\n  display: flex;\n}") {
		t.Errorf("expected a revealed Split to stay flex, got:\n%s", s)
	}
}

func TestVeilBlursWhatIsBehindIt(t *testing.T) {
	w := testWidget{name: "dlg", kind: widget.Dialog}
	s := style.For(w).Part("backdrop", style.Backdrop(style.Viewport), style.Veil()).Stylesheet().String()

	for _, want := range []string{
		"backdrop-filter: blur(" + css.VeilBlur.Var() + ");",
		"-webkit-backdrop-filter: blur(" + css.VeilBlur.Var() + ");",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("expected Veil to blur with %q, got:\n%s", want, s)
		}
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

// TestPrimarySurface_GradientHookIsInertByDefaultAndLiveWhenSet is the
// consumer-shaped proof for css.SetGradient (tinywasm/css) and the
// background-image line it depends on (style/emit_decls.go): a Primary
// surface always emits the hook, it costs nothing when no app opts in, and
// when an app's own Theme() call sets it, the SAME rule's output carries it
// through — the two packages actually compose, not just each in isolation.
func TestPrimarySurface_GradientHookIsInertByDefaultAndLiveWhenSet(t *testing.T) {
	w := testWidget{name: "w", kind: widget.Region}
	s := style.For(w).Part("cta", style.As(style.Primary)).Stylesheet().String()

	if !strings.Contains(s, "background-color: "+css.ColorPrimary.EnhancedVar()+";") {
		t.Errorf("expected the solid background-color to still be emitted, got:\n%s", s)
	}
	if !strings.Contains(s, "background-image: var(--color-primary-image, none);") {
		t.Errorf("expected the always-present, inert-by-default gradient hook, got:\n%s", s)
	}

	themed := css.Theme(
		css.Set(css.ColorPrimary, "#16a34a"),
		css.SetGradient(css.ColorPrimary, "135deg", css.ColorPrimary, css.ColorAccent),
	).String()

	if !strings.Contains(themed, "--color-primary-image: linear-gradient(135deg, var(--color-primary") {
		t.Errorf("expected the app's Theme() to set the same --color-primary-image the widget rule reads, got:\n%s", themed)
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
