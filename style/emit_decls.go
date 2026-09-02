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

	// as an inherited custom property: any Scroll() descendant — in this
	// widget or another — reserves it through var(--floating-bottom, 0px).
	if r.hasFloatingChrome {
		name := floatingBottomVar
		if r.floatingChromeEdge == EdgeTop {
			name = floatingTopVar
		}
		decls = append(decls, name+": calc("+iconSizeValue(r.floatingChromeSize)+" + 2 * "+spaceVar(r.floatingChromeGap)+");")
	}

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
		case flowFixedGrid:
			decls = append(decls, "--gap: "+spaceVar(r.flowGap)+";")
			decls = append(decls, "--cols: "+fmt.Sprint(r.flowCols)+";")
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

	decls = append(decls, r.surfaceDecls()...)

	// Interactive() is the DSL's own declaration that this part answers to
	// hover/focus/press — i.e. it is clickable by construction, not by a
	// per-widget styling choice. cursor: pointer is the mechanical consequence
	// of that flag, so it belongs here once instead of being repeated (and
	// inevitably forgotten somewhere) in every component that calls
	// Interactive().
	if r.interactive {
		decls = append(decls, "cursor: pointer;")
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
	if r.hasChipSeat {
		if r.chipSeatEdge == EdgeTop {
			decls = append(decls, "padding-block-start: calc(0.5 * "+css.ChipHeight.Var()+");")
		} else {
			decls = append(decls, "padding-block-end: calc(0.5 * "+css.ChipHeight.Var()+");")
		}
	}
	if r.hasPadInline {
		decls = append(decls, "padding-inline: "+spaceVar(r.padInline)+";")
	}
	if r.hasRound {
		decls = append(decls, "border-radius: "+radiusVar(r.round)+";")
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

	if r.hasRotate {
		decls = append(decls, "transform: rotate("+turnValue(r.rotate)+");")
	}

	decls = append(decls, r.placementDecls(layer)...)

	if r.hasRevealed && !r.hasDrawer {
		decls = append(decls, "display: none;")
	}

	if r.shown {
		decls = append(decls, "display: "+displayFor(r.flowType)+";")
	}

	if r.hidden {
		decls = append(decls, "display: none;")
	}

	return decls
}

func primitiveDecls(r rule) []string {
	var decls []string
	if r.fill || r.scroll {
		decls = append(decls, "height: 100%;", "min-height: 0;", "flex-grow: 1;")
	}
	if r.scroll {
		decls = append(decls, "overflow-y: auto;")
		// Every scroll region reserves the strip a FloatingChrome ancestor
		// declares it occupies — 0px when nobody declares one. ScrollGutter
		// folds its own ambient gutter into that same reservation.
		if r.hasScrollGutter {
			decls = append(decls, floatingPadDeclsWithGutter(r.scrollGutter)...)
		} else {
			decls = append(decls, floatingPadDecls()...)
		}
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
	if r.hasDivider {
		prop := "border-inline-end"
		if r.dividerSide == SideStart {
			prop = "border-inline-start"
		}
		decls = append(decls,
			prop+": "+borderStyle+css.ColorOutline.LightValue()+";",
			prop+": "+borderStyle+css.ColorOutline.EnhancedVar()+";",
		)
	}
	if r.hasDividerBelow {
		decls = append(decls,
			"border-block-end: "+borderStyle+css.ColorOutline.LightValue()+";",
			"border-block-end: "+borderStyle+css.ColorOutline.EnhancedVar()+";",
		)
	}
	return decls
}

func (r rule) emitsNothing(layer widget.Layer) bool {
	if len(r.Decls(layer)) > 0 {
		return false
	}
	return !r.hasFlow && !r.fill && !r.grow && !r.pushEnd && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow && !r.hasIcon && !r.controlBox && !r.logoBox && !r.chipBox && !r.hasGlyph && !r.hasPadEdge && !r.hasChipSeat && !r.hasPadInline && !r.startContent && !r.shown && !r.hasRotate
}

func formatRule(selectors []string, decls []string) string {
	if len(selectors) == 0 || len(decls) == 0 {
		return ""
	}
	sort.Strings(selectors)
	// Options overlap: Row and StartContent both say display:flex, and both are
	// legitimate on one rule — exact repeats must drop. This used to sort decls
	// first so identical strings landed adjacent, but alphabetical order is not
	// emission order: "padding-block-end:" sorts before "padding:" (- < : in
	// ASCII) regardless of which Option actually ran second, so Pad() followed
	// by PadEdge() — meant to override one edge on top of the general value —
	// silently came out with the shorthand LAST, winning over the longhand it
	// was supposed to lose to. A set-based dedup catches the same exact-string
	// repeats (now non-adjacent ones too, which the old approach missed)
	// without reordering anything: the CSS engine breaks equal-specificity
	// ties by source order, so the sequence Decls() appended in is the one
	// property override intent actually depends on.
	if len(decls) > 1 {
		seen := make(map[string]bool, len(decls))
		uniq := decls[:0]
		for _, d := range decls {
			if !seen[d] {
				seen[d] = true
				uniq = append(uniq, d)
			}
		}
		decls = uniq
	}
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
