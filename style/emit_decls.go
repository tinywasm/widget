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

	if r.hidden {
		decls = append(decls, "display: none;")
	}

	return decls
}

func (r rule) emitsNothing(layer widget.Layer) bool {
	if len(r.Decls(layer)) > 0 {
		return false
	}
	return !r.hasFlow && !r.fill && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow
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
