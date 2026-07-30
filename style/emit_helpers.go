//go:build !wasm

package style

import (
	"github.com/tinywasm/css"
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
	case flowSplit, flowGrid, flowFillCentered, flowSidebar:
		return "grid"
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
	default:
		return css.Token{}
	}
}
