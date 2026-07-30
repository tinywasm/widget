//go:build !wasm

package style

import (
	"sort"

	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func (r rule) Decls(layer widget.Layer) []string {
	var decls []string

	if r.hasFlow {
		switch r.flowType {
		case flowStack:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
		case flowRow:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
		case flowSplit:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
			decls = append(decls, "--ratio: "+splitRatioValue(r.flowRatio)+";")
		case flowGrid:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
			decls = append(decls, "--track: "+columnWidthValue(r.flowWidth)+";")
		case flowCenter:
			if r.hasSize {
				decls = append(decls, "--max-width: "+sizeValue(r.size)+";")
			} else {
				decls = append(decls, "--max-width: "+css.MaxWReadable.Var()+";")
			}
		case flowScrollRow:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
		case flowMediaBox:
			decls = append(decls, "--ratio: "+aspectValue(r.flowAspect)+";")
		case flowSidebar:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
			decls = append(decls, "--rail: "+railVar(r.flowRail)+";")
		}
	}

	if r.hasSurface {
		t := r.surface.resolve()
		if t.bg != "" {
			decls = append(decls, "background-color: "+t.bg+";")
		}
		if r.surface.defaultRadius() != RadiusNone && !r.hasRound {
			decls = append(decls, "border-radius: "+radiusVar(r.surface.defaultRadius())+";")
		}
		if t.text != "" {
			decls = append(decls, "color: "+t.text+";")
		}
		if t.border != "" {
			decls = append(decls, "border: "+t.border+";")
		}
	}

	if r.hasPad {
		decls = append(decls, "padding: "+spaceVar(r.pad)+";")
	}
	if r.hasPadEdge {
		if r.padEdge == EdgeTop {
			decls = append(decls, "padding-block-start: "+spaceVar(r.padEdgeSpace)+";")
		} else {
			decls = append(decls, "padding-block-end: "+spaceVar(r.padEdgeSpace)+";")
		}
	}
	if r.hasRound {
		decls = append(decls, "border-radius: "+radiusVar(r.round)+";")
	}
	if r.hasRaise {
		decls = append(decls, "box-shadow: "+elevationVar(r.raise)+";")
	}
	if r.hasSize {
		if !r.hasFlow || r.flowType != flowCenter {
			decls = append(decls, "width: "+sizeValue(r.size)+";")
		}
	}
	if r.hasIcon {
		v := iconSizeValue(r.icon)
		decls = append(decls, "width: "+v+";")
		decls = append(decls, "height: "+v+";")
		decls = append(decls, "flex-shrink: 0;")
	}
	if r.hasTextSize {
		decls = append(decls, "font-size: "+textSizeVar(r.textSize)+";")
	}
	if r.hasWeight {
		decls = append(decls, "font-weight: "+weightVar(r.weight)+";")
	}
	if r.hasMotion {
		decls = append(decls, "transition: "+motionValue(r.motion)+";")
	}

	if r.hasBackdrop {
		if r.backdropScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "inset: 0;")
		decls = append(decls, "z-index: "+layerVar(layer)+";")
	}

	if r.hasVeil {
		decls = append(decls, "background-color: color-mix(in srgb, "+css.ColorSurface.Var()+" 60%, transparent);")
	}

	if r.hasRevealed {
		decls = append(decls, "display: none;")
	}

	if r.hasDrawer {
		decls = append(decls, "position: fixed;")
		decls = append(decls, "inset-block: 0;")
		if r.drawerSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;")
		} else {
			decls = append(decls, "inset-inline-end: 0;")
		}
		decls = append(decls, "width: "+sizeValue(r.drawerSize)+";")
		decls = append(decls, "z-index: "+layerVar(layer)+";")
	}

	if r.hasGlyph {
		decls = append(decls, "color: "+familyBase(r.glyph).Var()+";")
		decls = append(decls, "fill: currentColor;")
	}

	if r.chipBox {
		decls = append(decls, "width: "+css.ChipWidth.Var()+";")
		decls = append(decls, "overflow: hidden;")
	}

	if r.controlBox {
		decls = append(decls, "min-height: "+css.ControlHeight.Var()+";")
	}

	if r.centerContent {
		decls = append(decls, "display: flex;")
		decls = append(decls, "align-items: center;")
		decls = append(decls, "justify-content: center;")
	}

	if r.hasAnchor {
		decls = append(decls, "position: relative;")
	}

	if r.hasDocked {
		if r.dockedScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "margin: 0;")
		if r.dockedEdge == EdgeTop {
			decls = append(decls, "inset-block-start: "+spaceVar(r.dockedGap)+";")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.dockedGap)+";")
		}
		if r.dockedSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.dockedGap)+";")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.dockedGap)+";")
		}
		decls = append(decls, "z-index: "+layerVar(layer)+";")
	}

	if r.hasOnEdge {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "margin: 0;")
		// translate by half of the element's OWN height: the straddle stays
		// exact whatever the chip's font size and padding turn out to be, which
		// a fixed negative margin can only approximate.
		if r.onEdgeEdge == EdgeTop {
			decls = append(decls, "inset-block-start: "+spaceVar(r.onEdgeBlock)+";")
			decls = append(decls, "transform: translateY(-50%);")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.onEdgeBlock)+";")
			decls = append(decls, "transform: translateY(50%);")
		}
		if r.onEdgeSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.onEdgeInline)+";")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.onEdgeInline)+";")
		}
		decls = append(decls, "z-index: "+layerVar(layer)+";")
	}

	if r.hasFlyout {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "inset-block-start: 100%;")
		if r.flyoutSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;")
		} else {
			decls = append(decls, "inset-inline-end: 0;")
		}
		decls = append(decls, "z-index: "+layerVar(layer)+";")
	}

	if r.hidden {
		decls = append(decls, "display: none;")
	}

	return decls
}

// primitiveDecls returns the declarations the boolean layout flags stand for.
// The main emission path groups them across selectors; a device-scoped rule has
// nothing to group with and emits them on its own selector.
func primitiveDecls(r rule) []string {
	var decls []string
	if r.fill || r.scroll {
		decls = append(decls, "height: 100%;", "min-height: 0;", "flex-grow: 1;")
	}
	if r.scroll {
		decls = append(decls, "overflow-y: auto;")
	}
	if r.grow {
		decls = append(decls, "flex-grow: 1;", "min-width: 0;")
	}
	if r.pushEnd {
		decls = append(decls, "margin-inline-start: auto;")
	}
	if r.keepSize {
		decls = append(decls, "flex-shrink: 0;", "flex-grow: 0;")
	}
	if r.edgeToEdge {
		decls = append(decls, "margin: 0;", "border-radius: 0;")
	}
	if r.hideOverflow {
		decls = append(decls, "overflow: hidden;")
	}
	return decls
}

func (r rule) emitsNothing(layer widget.Layer) bool {
	if len(r.Decls(layer)) > 0 {
		return false
	}
	return !r.hasFlow && !r.fill && !r.grow && !r.pushEnd && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow && !r.hasIcon && !r.controlBox && !r.chipBox && !r.hasGlyph && !r.hasPadEdge
}

func formatRule(selectors []string, decls []string) string {
	if len(selectors) == 0 || len(decls) == 0 {
		return ""
	}
	sort.Strings(selectors)
	sort.Strings(decls)
	sb := fmt.GetConv()
	defer sb.PutConv()
	sb.WriteString(fmt.JoinSlice(selectors, ", "))
	sb.WriteString(" {\n")
	for _, d := range decls {
		sb.WriteString("  ")
		sb.WriteString(d)
		sb.WriteString("\n")
	}
	sb.WriteString("}\n")
	return sb.GetString(fmt.BuffOut)
}
