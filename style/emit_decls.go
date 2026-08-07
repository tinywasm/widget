//go:build !wasm

package style

import (
	"sort"
	"strings"

	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func (r rule) Decls(layer widget.Layer) []string {
	var decls []string

	// FloatingChrome declares the strip this element occupies along its edge
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

	var t triplet
	if r.hasSurface {
		t = r.surface.resolve()
		if t.bg != "" {
			// Static first, enhanced second: a browser without light-dark()/
			// color-mix() support drops the second (invalid at parse time) and
			// keeps the first — permanently the light theme. A supporting
			// browser applies the second, which wins by being later.
			if t.bgStatic != "" {
				decls = append(decls, "background-color: "+t.bgStatic+";")
			}
			decls = append(decls, "background-color: "+t.bg+";")
		}
		// edgeToEdge's border-radius: 0 lives in the primitives layer, which the
		// widgets layer outranks — a default radius emitted here would win over
		// it and leave the box rounded against the frame (the crudview root
		// measured 4px with EdgeToEdge already applied). An explicit Round()
		// still wins: it is emitted in the widgets layer, same as the surface.
		if r.surface.defaultRadius() != RadiusNone && !r.hasRound && !r.edgeToEdge {
			decls = append(decls, "border-radius: "+radiusVar(r.surface.defaultRadius())+";")
		}
		if t.text != "" {
			if t.textStatic != "" {
				decls = append(decls, "color: "+t.textStatic+";")
			}
			decls = append(decls, "color: "+t.text+";")
			// Safari paints a :disabled control's text with its own
			// -webkit-text-fill-color, which WINS over color — a read-only
			// field styled for legibility still came out washed grey on iOS
			// while every other browser honoured the color above. Only state
			// rules pay for the extra pair: disabled/locked is a state, and a
			// surface's promise to resolve text together with its background
			// is not kept on the one platform where the text half is silently
			// overridden.
			if r.overlay {
				if t.textStatic != "" {
					decls = append(decls, "-webkit-text-fill-color: "+t.textStatic+";")
				}
				decls = append(decls, "-webkit-text-fill-color: "+t.text+";")
			}
		}
		if t.border != "" && !r.overlay {
			// A state rule repaints its border as a shadow ring through
			// boxShadowDecls below — never as a border, which would grow the
			// box exactly when the pointer entered it.
			if t.borderStatic != "" {
				decls = append(decls, "border: "+t.borderStatic+";")
			}
			decls = append(decls, "border: "+t.border+";")
		}
	}

	// The one box-shadow decision point: Raise() and a state border both paint
	// through box-shadow, and the two compose here — ring first, elevation
	// after — or the later declaration would silently stomp the earlier one.
	decls = append(decls, boxShadowDecls(r, t)...)

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

	if r.hasBackdrop {
		if r.backdropScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "inset: 0;")
		// The stacking level is stackingFor's call — the single site that
		// decides local chrome from real overlay.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasVeil {
		decls = append(decls, "background-color: "+css.FadeStatic(css.ColorSurface, 0.4)+";")
		decls = append(decls, "background-color: color-mix(in srgb, "+css.ColorSurface.NestedEnhanced()+" 60%, transparent);")
		// A wash alone still lets the page behind compete for attention.
		// Softening it is what makes the thing on top read as the only thing in
		// focus. -webkit- first: Safari has never unprefixed this.
		decls = append(decls, "-webkit-backdrop-filter: blur("+css.VeilBlur.Var()+");")
		decls = append(decls, "backdrop-filter: blur("+css.VeilBlur.Var()+");")
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
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasGlyph {
		decls = append(decls, "color: "+familyBase(r.glyph).LightValue()+";")
		decls = append(decls, "color: "+familyBase(r.glyph).EnhancedVar()+";")
		decls = append(decls, "fill: currentColor;")
	}

	if r.chipBox {
		decls = append(decls, "width: "+css.ChipWidth.Var()+";")
		decls = append(decls, "overflow: hidden;")
	}

	if r.capitalize {
		decls = append(decls, "text-transform: capitalize;")
	}

	if r.controlBox {
		decls = append(decls, "min-height: "+css.ControlHeight.Var()+";")
	}

	if r.centerContent {
		decls = append(decls, "display: flex;")
		decls = append(decls, "align-items: center;")
		decls = append(decls, "justify-content: center;")
	}

	if r.startContent {
		decls = append(decls, "display: flex;")
		decls = append(decls, "align-items: center;")
		decls = append(decls, "justify-content: flex-start;")
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
		// All four insets, the unpinned pair explicitly auto. A corner pin owns
		// the whole box: leaving the opposite edges unset lets whatever set them
		// before survive, and an over-constrained absolute box collapses instead
		// of moving. That is what happened when this overrode a Flyout on one
		// device — top and bottom were both live and the panel became 8x2.
		if r.dockedEdge == EdgeTop {
			decls = append(decls, "inset-block-start: "+spaceVar(r.dockedGap)+";", "inset-block-end: auto;")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.dockedGap)+";", "inset-block-start: auto;")
		}
		if r.dockedSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.dockedGap)+";", "inset-inline-end: auto;")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.dockedGap)+";", "inset-inline-start: auto;")
		}
		// Parent vs Viewport docking levels are stackingFor's call: a Parent
		// dock is local chrome, a Viewport dock claims the widget's layer.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasOnEdge {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "margin: 0;")
		// Half a chip-height of negative margin, not a transform: the box the
		// chip straddles now EXISTS for scrollHeight, which a transform is
		// invisible to by spec — an ancestor could never reserve the space the
		// chip actually occupies. The straddle is exact because the chip's
		// height is the shared --chip-height token (the margin was chosen to be
		// height-agnostic only while the token did not exist); the fallback
		// keeps the behavior against an older token sheet. The transform also
		// created an implicit stacking context; the margin creates none.
		if r.onEdgeEdge == EdgeTop {
			decls = append(decls, "inset-block-start: "+spaceVar(r.onEdgeBlock)+";")
			decls = append(decls, "margin-block-start: calc(-0.5 * "+css.ChipHeight.Var()+");")
		} else {
			decls = append(decls, "inset-block-end: "+spaceVar(r.onEdgeBlock)+";")
			decls = append(decls, "margin-block-end: calc(-0.5 * "+css.ChipHeight.Var()+");")
		}
		if r.onEdgeSide == SideStart {
			decls = append(decls, "inset-inline-start: "+spaceVar(r.onEdgeInline)+";")
		} else {
			decls = append(decls, "inset-inline-end: "+spaceVar(r.onEdgeInline)+";")
		}
		// Local level, never the overlay layer — stackingFor is the single
		// source of that decision.
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hasFlyout {
		decls = append(decls, "position: absolute;")
		decls = append(decls, "inset-block-start: 100%;", "inset-block-end: auto;")
		if r.flyoutSide == SideStart {
			decls = append(decls, "inset-inline-start: 0;", "inset-inline-end: auto;")
		} else {
			decls = append(decls, "inset-inline-end: 0;", "inset-inline-start: auto;")
		}
		if z := stackingFor(r, layer); z != "" {
			decls = append(decls, z)
		}
	}

	if r.hidden {
		decls = append(decls, "display: none;")
	}

	return decls
}

// primitiveDecls returns the declarations the boolean layout flags stand for.
// The main emission path groups them across selectors; a device-scoped rule has
// nothing to group with and emits them on its own selector.
// boxShadowDecls is the only place in the package that emits box-shadow,
// composing Raise()'s elevation with a state border's ring when both land on
// one rule — ring first, elevation after, in a single declaration, so neither
// stomps the other.
//
// A state border is a shadow ring (0 0 0 1px <color>) because an outline was
// a two-fold defect: outlines paint at the END of the stacking context, over
// the element's positioned descendants (a Locked border crossed over the
// legend riding the same line), and Safari < 16.4 ignores border-radius on
// outlines — every state border in the system rendered square. The ring
// paints with the element's own background, below its descendants, on a
// radius every engine honors, and it keeps the original promise: no layout
// space, so the element does not grow under the pointer that entered it.
func boxShadowDecls(r rule, t triplet) []string {
	if t.border != "" && r.overlay {
		if r.hasRaise {
			// One declaration, and the ring's color is the token's Var() form,
			// not the static+enhanced pair: the elevation's var() inside the
			// light-dark() half would defer the WHOLE declaration to
			// computed-value time (the parse-time-safe rule), silently
			// discarding the static half on a browser without light-dark().
			// A bare var() with its hex fallback is safe on every engine: a
			// legacy browser cannot compute --color-outline's light-dark()
			// :root value, so the fallback applies.
			return []string{"box-shadow: " + ringShadow(t.borderVar) + ", " + elevationVar(r.raise) + ";"}
		}
		if t.borderStatic != "" {
			// Static ring first, enhanced ring second — the same double
			// declaration every other themed color uses.
			return []string{
				"box-shadow: " + ringShadow(t.borderStatic) + ";",
				"box-shadow: " + ringShadow(t.border) + ";",
			}
		}
		return []string{"box-shadow: " + ringShadow(t.border) + ";"}
	}
	if r.hasRaise {
		return []string{"box-shadow: " + elevationVar(r.raise) + ";"}
	}
	return nil
}

// ringShadow turns a border value ("1px solid <color>") into the shadow ring
// a state border paints with: 0 0 0 1px <color>.
func ringShadow(border string) string {
	return "0 0 0 1px " + strings.TrimPrefix(border, borderStyle)
}

func primitiveDecls(r rule) []string {
	var decls []string
	if r.fill || r.scroll {
		decls = append(decls, "height: 100%;", "min-height: 0;", "flex-grow: 1;")
	}
	if r.scroll {
		decls = append(decls, "overflow-y: auto;")
		// Every scroll region reserves the strip a FloatingChrome ancestor
		// declares it occupies — 0px when nobody declares one.
		decls = append(decls, floatingPadDecls()...)
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
	return !r.hasFlow && !r.fill && !r.grow && !r.pushEnd && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow && !r.hasIcon && !r.controlBox && !r.chipBox && !r.hasGlyph && !r.hasPadEdge && !r.hasPadInline && !r.startContent
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
