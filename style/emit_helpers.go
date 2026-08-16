//go:build !wasm

package style

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func selectorOf(name widget.Name, part widget.Part) string {
	if part == "" {
		return "." + string(name)
	}
	return "." + string(name) + "__" + string(part)
}

func cuePseudo(c widget.Cue) string {
	switch c {
	case widget.Hover:
		return ":hover"
	case widget.Focus:
		return ":focus-visible"
	case widget.Press:
		return ":active"
	case widget.Target:
		return ":target"
	default:
		return ""
	}
}

func spaceVar(s Space) string {
	switch s {
	case SpaceNone:
		return "0"
	case Space1:
		return css.Space1.Var()
	case Space2:
		return css.Space2.Var()
	case Space3:
		return css.Space3.Var()
	case Space4:
		return css.Space4.Var()
	case Space6:
		return css.Space6.Var()
	case Space8:
		return css.Space8.Var()
	case Space12:
		return css.Space12.Var()
	default:
		return "0"
	}
}

func radiusVar(r Radius) string {
	switch r {
	case RadiusNone:
		return "0"
	case RadiusSm:
		return css.RadiusSm.Var()
	case RadiusMd:
		return css.RadiusMd.Var()
	case RadiusLg:
		return css.RadiusLg.Var()
	case RadiusFull:
		return css.RadiusFull.Var()
	default:
		return "0"
	}
}

func elevationVar(e Elevation) string {
	switch e {
	case Flat:
		return "none"
	case Raised:
		return css.ShadowSm.Var()
	case Floating:
		return css.ShadowMd.Var()
	case Popover:
		return css.ShadowLg.Var()
	default:
		return "none"
	}
}

// splitRatioValue feeds flex-grow, which takes a unitless number. An fr unit
// here is invalid at computed-value time, which silently resets flex-grow to
// its initial 0 and leaves the first partition at its content width.
func splitRatioValue(r SplitRatio) string {
	switch r {
	case SplitHalf:
		return "1"
	case SplitTwoThirds:
		return "2"
	case SplitThreeQuarters:
		return "3"
	default:
		return "1"
	}
}

func aspectValue(a Aspect) string {
	switch a {
	case AspectSquare:
		return "1/1"
	case Aspect3x2:
		return "3/2"
	case Aspect4x3:
		return "4/3"
	case Aspect16x9:
		return "16/9"
	default:
		return "1/1"
	}
}

func columnWidthValue(cw ColumnWidth) string {
	switch cw {
	case ColumnNarrow:
		return css.ColumnNarrow.Var()
	case ColumnMedium:
		return css.ColumnMedium.Var()
	case ColumnWide:
		return css.ColumnWide.Var()
	default:
		return css.ColumnNarrow.Var()
	}
}

func sizeValue(s Size) string {
	switch s {
	case Content:
		return "max-content"
	case Readable:
		return css.MaxWReadable.Var()
	case Compact:
		// A stack of controls is not prose: Readable's 65ch is the measure a
		// LINE of text wants, and a login card or a settings pane sized to it
		// reads as a form someone forgot to constrain. 24rem is the width a
		// single column of fields is legible at.
		//
		// min(), unlike every other step here, because this one is a cap and
		// not a share: the percentages resolve against the container and shrink
		// with it for free, while a bare 24rem would overflow a phone. The
		// element takes the container up to 24rem and stops.
		return "min(100%, 24rem)"
	case Third:
		return "33.33%"
	case Half:
		return "50%"
	case TwoThirds:
		return "66.66%"
	case Most:
		return "90%"
	case Full:
		return "100%"
	default:
		return "auto"
	}
}

func iconSizeValue(s IconSize) string {
	switch s {
	case IconSm:
		return "1em"
	case IconLg:
		return "2.5em"
	default:
		return "1.5em"
	}
}

func textSizeVar(ts TextSize) string {
	switch ts {
	case TextXs:
		return css.TextXs.Var()
	case TextSm:
		return css.TextSm.Var()
	case TextBase:
		return css.TextBase.Var()
	case TextLg:
		return css.TextLg.Var()
	case TextXl:
		return css.TextXl.Var()
	case Text2xl:
		return css.Text2xl.Var()
	default:
		return "inherit"
	}
}

func weightVar(w Weight) string {
	switch w {
	case WeightRegular:
		return css.FontWeightRegular.Var()
	case WeightMedium:
		return css.FontWeightMedium.Var()
	case WeightBold:
		return css.FontWeightBold.Var()
	default:
		return "inherit"
	}
}

func motionValue(m Motion) string {
	switch m {
	case MotionNone:
		return "none"
	case MotionFast:
		return "all " + css.DurationFast.Var() + " " + css.EaseInOut.Var()
	case MotionBase:
		return "all " + css.DurationBase.Var() + " " + css.EaseInOut.Var()
	case MotionSlow:
		return "all " + css.DurationSlow.Var() + " " + css.EaseInOut.Var()
	default:
		return "none"
	}
}

func displayFor(f flowType) string {
	switch f {
	case flowStack, flowRow, flowScrollRow, flowMediaBox, flowCover, flowMasterDetail:
		return "flex"
	case flowGrid, flowFillCentered:
		return "grid"
	case flowSplit, flowSidebar:
		// Both emit display:flex in @layer primitives. Saying "grid" here would
		// win from @layer states and strand every flex-basis/flex-grow the
		// primitive laid down.
		return "flex"
	default:
		return "block"
	}
}

func layerVar(l widget.Layer) string {
	switch l {
	case widget.LayerBase:
		return css.ZBase.Var()
	case widget.LayerDropdown:
		return css.ZDropdown.Var()
	case widget.LayerSticky:
		return css.ZSticky.Var()
	case widget.LayerModal:
		return css.ZModal.Var()
	case widget.LayerToast:
		return css.ZToast.Var()
	case widget.LayerTooltip:
		return css.ZTooltip.Var()
	default:
		return css.ZBase.Var()
	}
}

// stackingFor is the package's ONLY source of a z-index value. An element
// taken out of the flow must never be left at `auto`: auto means "whatever
// the DOM order and the engine's compositor happen to do" — the silent
// failure of the OnEdge-under-the-input bug, where Safari composites a UA
// form control over an auto-z-index sibling that every other engine painted
// on top. Two declared levels exist:
//
//   - local (1): chrome that rides on its own widget's content — OnEdge, and
//     Parent-scoped Docked/Backdrop/EdgeStrip. Deliberately NOT the overlay
//     layer: overlay tokens start at --z-dropdown (100+), and a chip level
//     with a real dropdown would tie it and win on DOM order, rendering
//     under it. The same applies to a Parent backdrop: on the overlay layer
//     a dialog's click-catcher outranked its own panel, and the blur it
//     carried blurred the very thing it was meant to isolate.
//   - overlay (var(--z-dropdown) and up, via Kind.Layer()): real overlays —
//     Flyout, Drawer, Backdrop(Viewport), Docked(Viewport),
//     EdgeStrip(Viewport).
//
// Returns "" only when nothing in the rule is out of flow; the emitter then
// emits nothing, and Validate() flags an out-of-flow rule whose level came
// back unresolvable.
func stackingFor(r rule, layer widget.Layer) string {
	switch {
	case r.hasOnEdge:
		return "z-index: 1;"
	case r.hasDocked && r.dockedScope == Parent:
		return "z-index: 1;"
	case r.hasEdgeStrip && r.edgeStripScope == Parent:
		return "z-index: 1;"
	case r.hasBackdrop && r.backdropScope == Parent:
		return "z-index: 1;"
	case r.hasDrawer:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasDocked && r.dockedScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasEdgeStrip && r.edgeStripScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasFlyout:
		return "z-index: " + layerVar(layer) + ";"
	case r.hasBackdrop && r.backdropScope == Viewport:
		return "z-index: " + layerVar(layer) + ";"
	default:
		return ""
	}
}

// The custom properties the floating-chrome seam is built on: a FloatingChrome
// host declares the strip it occupies along its edge, and every Scroll()
// region descendant reserves it through var(--floating-<edge>, 0px). They are
// DSL-owned variables crossing a component boundary, so they are not css
// tokens — the drift guard knows them by name.
const (
	floatingTopVar    = "--floating-top"
	floatingBottomVar = "--floating-bottom"
)

// floatingPadDecls are the reservations a Scroll() region carries: whatever a
// FloatingChrome ancestor declares it occupies along the region's edges —
// nothing declared means 0px, which is the default padding anyway, so the
// cost of the seam for authors who do not use it is zero.
func floatingPadDecls() []string {
	return []string{
		"padding-block-start: var(" + floatingTopVar + ", 0px);",
		"padding-block-end: var(" + floatingBottomVar + ", 0px);",
	}
}

func railVar(rw RailWidth) string {
	switch rw {
	case RailNarrow:
		return css.RailNarrow.Var()
	case RailWide:
		return css.RailWide.Var()
	default:
		return css.RailNarrow.Var()
	}
}

// flowSelfDecls returns the declarations a flow puts on the CONTAINER itself,
// without the child rules that go with it. The main path groups flows into
// shared selectors collected from the root and the parts; a rule that is
// neither — a CueWithin — has nothing to group with and needs them inline, or
// it silently emits no `display` at all.
func flowSelfDecls(r rule) []string {
	switch r.flowType {
	case flowStack:
		return []string{"display: flex;", "flex-direction: column;", "gap: var(--gap);", "min-height: 0;"}
	case flowRow:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);", "align-items: center;"}
	case flowSplit:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}
	case flowGrid:
		return []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));"}
	case flowFixedGrid:
		return []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(var(--cols), minmax(0, 1fr));"}
	case flowCenter:
		return []string{"margin-inline: auto;", "max-width: var(--max-width);", "width: 100%;"}
	case flowFillCentered:
		return []string{"display: grid;", "place-items: center;", "min-height: 100%;", "width: 100%;"}
	case flowScrollRow:
		return []string{"display: flex;", "gap: var(--gap);", "overflow-x: auto;", "scroll-snap-type: x mandatory;", "scroll-behavior: smooth;"}
	case flowMediaBox:
		return []string{"aspect-ratio: var(--ratio);", "overflow: hidden;", "display: flex;", "justify-content: center;", "align-items: center;"}
	case flowCover:
		return []string{"height: 100dvh;", "display: flex;", "flex-direction: column;"}
	case flowSidebar:
		return []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}
	case flowSlideDeck:
		return slideDeckStripDecls()
	case flowMasterDetail:
		return masterDetailStripDecls()
	default:
		return nil
	}
}

// motionDurationVar es solo la duración de un Motion, sin propiedad ni easing.
func motionDurationVar(m Motion) string {
	switch m {
	case MotionFast:
		return css.DurationFast.Var()
	case MotionBase:
		return css.DurationBase.Var()
	case MotionSlow:
		return css.DurationSlow.Var()
	default:
		return "0s"
	}
}

// slideDeckStripDecls: el contenedor es el bloque contenedor de las capas y
// recorta lo que queda aparcado afuera.
func slideDeckStripDecls() []string {
	return []string{
		"position: relative;",
		"overflow: hidden;",
	}
}

// slideDeckPageDecls: cada hijo cubre el contenedor y espera aparcado en el borde
// inline-start. visibility (no display) es lo que lo saca del orden de tabulación
// sin impedir la transición: se apaga DESPUÉS de que el deslizamiento termina, por
// eso el retardo en la segunda propiedad.
func slideDeckPageDecls(m Motion) []string {
	d := motionDurationVar(m)
	return []string{
		"position: absolute;",
		"inset: 0;",
		"transform: translateX(-100%);",
		"visibility: hidden;",
		"transition: transform " + d + " " + css.EaseInOut.Var() + ", visibility 0s linear " + d + ";",
	}
}

// slideDeckCurrentDecls: el panel activo entra hasta su sitio y es visible desde
// el primer fotograma, para que se le vea llegar.
func slideDeckCurrentDecls(m Motion) []string {
	d := motionDurationVar(m)
	return []string{
		"transform: translateX(0);",
		"visibility: visible;",
		"transition: transform " + d + " " + css.EaseInOut.Var() + ", visibility 0s;",
	}
}

// autoRotateStepSeconds is how long one layer holds the screen before the
// next takes over. Not a token: tinywasm/css's duration scale (150-400ms) is
// for UI transitions, and an unattended background rotation runs one to two
// orders of magnitude slower than that — there is nothing in the catalog to
// reuse here, so this is a fixed implementation constant of AutoRotate, the
// same way slideDeckPageDecls' "0s" fallback above is.
const autoRotateStepSeconds = 5

// autoRotateCycleSeconds is the time for every layer to get exactly one
// turn. AutoRotateLayers is fixed (see its doc comment), so this is fixed
// too: every AutoRotate() rule in a page shares one @keyframes definition
// and one cycle length, regardless of how many of the layers a given
// instance's markup actually fills.
const autoRotateCycleSeconds = autoRotateStepSeconds * AutoRotateLayers

// autoRotateKeyframesName is the single @keyframes identifier every
// AutoRotate() rule references. Keyframe names are global in CSS, not
// scoped to a selector, and the crossfade shape below never varies, so one
// definition — duplicated harmlessly if more than one AutoRotate part
// appears in the same stylesheet — is simpler than inventing a per-instance
// name.
const autoRotateKeyframesName = "tw-auto-rotate"

// autoRotateStripDecls: the container is the containing block for its
// layered children and clips whatever a layer's transform pushes outside it.
func autoRotateStripDecls() []string {
	return []string{
		"position: relative;",
		"overflow: hidden;",
	}
}

// autoRotateLayerDecls: every layer covers the container and starts hidden.
// The animation is what makes each one visible in turn; a layer whose delay
// has not elapsed yet, or whose slot has no element mounted in it, simply
// keeps showing this resting opacity — which is why :first-child overrides
// it below, so the reduced-motion fallback (animation: none, see the
// prefers-reduced-motion block this rule feeds into) is never a blank layer.
func autoRotateLayerDecls() []string {
	return []string{
		"position: absolute;",
		"inset: 0;",
		"opacity: 0;",
		fmt.Sprintf("animation: %s %ds infinite;", autoRotateKeyframesName, autoRotateCycleSeconds),
	}
}

// autoRotateFirstDecls: the resting (non-animated) opacity for the first
// layer. It is redundant while the animation runs — animation always wins
// over a plain opacity declaration on the same property — and it is the
// entire reduced-motion fallback once the animation is switched off.
func autoRotateFirstDecls() []string {
	return []string{"opacity: 1;"}
}

// autoRotateDelayDecls staggers slot's turn by slot * autoRotateStepSeconds.
// slot is the layer's 1-based DOM position (1 is :first-child, which needs
// no delay — its turn is the animation's own 0%).
func autoRotateDelayDecls(slot int) []string {
	return []string{fmt.Sprintf("animation-delay: %ds;", (slot-1)*autoRotateStepSeconds)}
}

// autoRotateKeyframesCSS is the shared crossfade shape every layer plays out
// on its own staggered clock: fully visible for the first chunk of its slot,
// a short fade, then hidden for the rest of the cycle until its next turn.
// The 12%/16% breakpoints are fixed because autoRotateCycleSeconds is fixed
// — see AutoRotateLayers' doc comment for why this package cannot compute
// them from a real image count.
func autoRotateKeyframesCSS() string {
	return "@keyframes " + autoRotateKeyframesName + " {\n" +
		"  0%, 12% { opacity: 1; }\n" +
		"  16%, 100% { opacity: 0; }\n" +
		"}\n\n"
}

// masterDetailStripDecls lays the two panels out as a horizontal scroll-snap
// strip. direction: rtl is load-bearing — it puts the start edge on the right,
// which is where scroll position 0 already rests, so the master panel is what
// shows on arrival with no scroll nudge at mount time.
func masterDetailStripDecls() []string {
	return []string{
		"display: flex;",
		"flex-direction: row;",
		"flex-wrap: nowrap;",
		"direction: rtl;",
		"gap: 0;",
		"overflow-x: auto;",
		"overflow-y: hidden;",
		"scroll-snap-type: x mandatory;",
		"scroll-behavior: smooth;",
	}
}

// masterDetailResetDecls clears whatever the wide-screen flow left on the
// children. Split in particular gives every child a flex-basis of
// calc((40rem - 100%) * 999), which below 40rem is a six-figure pixel value: any
// child the two panel rules do not cover — a modal mount point, a portal anchor
// — keeps it and blows the strip's scroll width apart.
func masterDetailResetDecls() []string {
	return []string{
		"flex: 0 0 auto;",
		// direction: ltr on EVERY child, not just the two panels. The strip is
		// laid out rtl so the master lands at scroll position 0; a third child —
		// a modal mount point, a portal anchor — would otherwise inherit it and
		// render its text backwards.
		"direction: ltr;",
	}
}

// masterDetailDetailDecls sizes the detail panel. It is first in the DOM, the
// order a desktop Split wants, and order: 2 moves it beside the master without
// touching the markup. Its width is a share of the SCROLL CONTAINER, not of the
// viewport: the host panel is not guaranteed to be viewport-wide, and a vw here
// overflows the strip by the difference.
func masterDetailDetailDecls(detail Size) []string {
	return []string{
		"direction: ltr;",
		"flex: 0 0 " + sizeValue(detail) + ";",
		"scroll-snap-align: end;",
		"order: 2;",
	}
}

func masterDetailMasterDecls() []string {
	return []string{
		"direction: ltr;",
		"flex: 0 0 100%;",
		"scroll-snap-align: start;",
		"order: 1;",
	}
}

func sidebarRailSel(sel string, side Side) string {
	if side == SideStart {
		return sel + " > :first-child"
	}
	return sel + " > :last-child"
}

func sidebarContentSel(sel string, side Side) string {
	if side == SideStart {
		return sel + " > :last-child"
	}
	return sel + " > :first-child"
}

func familyBase(s Surface) css.Token {
	switch s {
	case Panel, Inset, Highlight:
		return css.ColorSurface
	case Primary:
		return css.ColorPrimary
	case Accent:
		return css.ColorAccent
	case Secondary:
		return css.ColorSurface
	case Success:
		return css.ColorSuccess
	case Danger:
		return css.ColorDanger
	case Subtle:
		return css.ColorMuted
	case Page:
		// Page is the whitest surface: a white base lets Interactive(Page)
		// derive a hover/focus/press family from the page background itself.
		return css.ColorBackground
	default:
		return css.Token{}
	}
}
