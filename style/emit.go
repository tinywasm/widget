//go:build !wasm

package style

import (
	"sort"

	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

// Selector helper
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
		return ":focus"
	case widget.Press:
		return ":active"
	case widget.Target:
		return ":target"
	default:
		return ""
	}
}

// Map enums to CSS variables with fallbacks using css package tokens
func spaceVar(s Space) string {
	switch s {
	case Space0:
		return "0"
	case Space1:
		return css.Space1.Var()
	case Space2:
		return css.Space2.Var()
	case Space3:
		return css.Space3.Var()
	case Space4:
		return css.Space4.Var()
	case Space5, Space6:
		return css.Space6.Var()
	case Space7, Space8:
		return css.Space8.Var()
	case Space9, Space10, Space11, Space12:
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
	case Overlay:
		return css.ShadowLg.Var()
	default:
		return "none"
	}
}

func ratioValue(r Ratio) string {
	switch r {
	case RatioHalf:
		return "1fr"
	case RatioTwoThirds:
		return "2fr"
	case RatioThreeQuarters:
		return "3fr"
	default:
		return "1fr"
	}
}

func ratioFraction(r Ratio) string {
	switch r {
	case RatioHalf:
		return "1/1"
	case RatioTwoThirds:
		return "3/2"
	case RatioThreeQuarters:
		return "4/3"
	default:
		return "1/1"
	}
}

func trackValue(t Track) string {
	switch t {
	case TrackSm:
		return "15rem"
	case TrackMd:
		return "25rem"
	case TrackLg:
		return "40rem"
	default:
		return "15rem"
	}
}

func sizeValue(s Size) string {
	switch s {
	case Content:
		return "max-content"
	case Prose:
		return css.MaxWProse.Var()
	case Third:
		return "33.33%"
	case Half:
		return "50%"
	case TwoThirds:
		return "66.66%"
	case Full:
		return "100%"
	default:
		return "auto"
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

// Rule.Decls formats the declarations of a rule
func (r Rule) Decls() []string {
	var decls []string

	// Flow custom variables
	if r.HasFlow {
		switch r.FlowType {
		case "stack":
			decls = append(decls, "--gap: "+spaceVar(r.FlowGap)+";")
		case "row":
			decls = append(decls, "--gap: "+spaceVar(r.FlowGap)+";")
		case "split":
			decls = append(decls, "--gap: "+spaceVar(r.FlowGap)+";")
			decls = append(decls, "--ratio: "+ratioValue(r.FlowRatio)+";")
		case "grid":
			decls = append(decls, "--gap: "+spaceVar(r.FlowGap)+";")
			decls = append(decls, "--track: "+trackValue(r.FlowTrack)+";")
		case "center":
			if r.HasSize {
				decls = append(decls, "--max-width: "+sizeValue(r.Size)+";")
			} else {
				decls = append(decls, "--max-width: "+css.MaxWProse.Var()+";")
			}
		case "reel":
			decls = append(decls, "--gap: "+spaceVar(r.FlowGap)+";")
		case "frame":
			decls = append(decls, "--ratio: "+ratioFraction(r.FlowRatio)+";")
		}
	}

	// Surface colors
	if r.HasSurface {
		t := r.Surface.Resolve()
		if t.Bg != "" {
			decls = append(decls, "background-color: "+t.Bg+";")
		}
		if t.Text != "" {
			decls = append(decls, "color: "+t.Text+";")
		}
		if t.Border != "" {
			decls = append(decls, "border: "+t.Border+";")
		}
	}

	// Scales
	if r.HasPad {
		decls = append(decls, "padding: "+spaceVar(r.Pad)+";")
	}
	if r.HasRound {
		decls = append(decls, "border-radius: "+radiusVar(r.Round)+";")
	}
	if r.HasRaise {
		decls = append(decls, "box-shadow: "+elevationVar(r.Raise)+";")
	}
	if r.HasSize {
		if !r.HasFlow || r.FlowType != "center" {
			decls = append(decls, "width: "+sizeValue(r.Size)+";")
		}
	}
	if r.HasTextSize {
		decls = append(decls, "font-size: "+textSizeVar(r.TextSize)+";")
	}
	if r.HasWeight {
		decls = append(decls, "font-weight: "+weightVar(r.Weight)+";")
	}

	// Backdrop / stacking / visibility declarations
	if r.HasBackdrop {
		if r.BackdropScope == Viewport {
			decls = append(decls, "position: fixed;")
		} else {
			decls = append(decls, "position: absolute;")
		}
		decls = append(decls, "inset: 0;")
		decls = append(decls, "z-index: 100;")
	}

	if r.Above {
		decls = append(decls, "z-index: 101;")
	}

	if r.Scrim {
		decls = append(decls, "background-color: color-mix(in srgb, var(--color-surface) 60%, transparent);")
	}

	if r.Hidden {
		decls = append(decls, "display: none;")
	}

	if r.Shown {
		decls = append(decls, "display: block;")
	}

	return decls
}

// Formats a CSS rule block deterministically.
func formatRule(selectors []string, decls []string) string {
	if len(selectors) == 0 || len(decls) == 0 {
		return ""
	}
	sort.Strings(selectors)
	sort.Strings(decls)
	sb := fmt.Convert()
	sb.WriteString(fmt.Convert(selectors).Join(", ").String() + " {\n")
	for _, d := range decls {
		sb.WriteString("  " + d + "\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

// Stylesheet genera un css.Stylesheet para el Sheet.
func (s *Sheet) Stylesheet() *css.Stylesheet {
	sb := fmt.Convert()

	// Fixed layer definition
	sb.WriteString("@layer tokens, primitives, widgets, states;\n\n")

	// 1. Gather selectors for primitives
	var stackSel, rowSel, splitSel, gridSel, centerSel, coverSel, reelSel, frameSel []string
	var fillSel, scrollsSel, fixedSel, flushSel, clipSel []string

	collect := func(r Rule, sel string) {
		if r.HasFlow {
			switch r.FlowType {
			case "stack":
				stackSel = append(stackSel, sel)
			case "row":
				rowSel = append(rowSel, sel)
			case "split":
				splitSel = append(splitSel, sel)
			case "grid":
				gridSel = append(gridSel, sel)
			case "center":
				centerSel = append(centerSel, sel)
			case "cover":
				coverSel = append(coverSel, sel)
			case "reel":
				reelSel = append(reelSel, sel)
			case "frame":
				frameSel = append(frameSel, sel)
			}
		}
		if r.Fill {
			fillSel = append(fillSel, sel)
		}
		if r.Scrolls {
			scrollsSel = append(scrollsSel, sel)
		}
		if r.Fixed {
			fixedSel = append(fixedSel, sel)
		}
		if r.Flush {
			flushSel = append(flushSel, sel)
		}
		if r.Clip {
			clipSel = append(clipSel, sel)
		}
	}

	collect(s.RootRule, selectorOf(s.Name, ""))

	// Sort parts to ensure deterministic gathering
	var parts []string
	for p := range s.PartRules {
		parts = append(parts, string(p))
	}
	sort.Strings(parts)

	for _, p := range parts {
		collect(s.PartRules[widget.Part(p)], selectorOf(s.Name, widget.Part(p)))
	}

	// 2. Emit @layer primitives
	sb.WriteString("@layer primitives {\n")

	emitPrimitive := func(base string, extraSel []string, decls []string) {
		if len(extraSel) > 0 {
			selectors := append([]string{base}, extraSel...)
			sb.WriteString(formatRule(selectors, decls))
		}
	}

	emitPrimitive(".fl-stack", stackSel, []string{
		"display: flex;",
		"flex-direction: column;",
		"min-height: 0;",
	})
	if len(stackSel) > 0 {
		var stackKids []string
		for _, sel := range stackSel {
			stackKids = append(stackKids, sel+" > * + *")
		}
		emitPrimitive(".fl-stack > * + *", stackKids, []string{
			"margin-block-start: var(--gap);",
		})
	}

	emitPrimitive(".fl-row", rowSel, []string{
		"display: flex;",
		"flex-wrap: wrap;",
		"gap: var(--gap);",
		"align-items: center;",
	})

	emitPrimitive(".fl-split", splitSel, []string{
		"container-type: inline-size;",
		"display: grid;",
		"gap: var(--gap);",
		"grid-template-columns: var(--ratio) 1fr;",
	})
	if len(splitSel) > 0 {
		sort.Strings(splitSel)
		sb.WriteString("@container (max-width: 40rem) {\n")
		var splitInner []string
		for _, sel := range splitSel {
			splitInner = append(splitInner, sel)
		}
		ruleStr := formatRule(splitInner, []string{
			"grid-template-columns: 1fr;",
		})
		indented := "  " + fmt.Convert([]string{ruleStr}).Join("\n  ").String()
		sb.WriteString(indented + "\n}\n")
	}

	emitPrimitive(".fl-grid", gridSel, []string{
		"display: grid;",
		"gap: var(--gap);",
		"grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));",
	})

	emitPrimitive(".fl-center", centerSel, []string{
		"margin-inline: auto;",
		"max-width: var(--max-width, var(--max-w-prose));",
		"width: 100%;",
	})

	emitPrimitive(".fl-cover", coverSel, []string{
		"display: grid;",
		"place-items: center;",
		"min-height: 100%;",
		"width: 100%;",
	})

	emitPrimitive(".fl-reel", reelSel, []string{
		"display: flex;",
		"gap: var(--gap);",
		"overflow-x: auto;",
		"scroll-snap-type: x mandatory;",
	})
	if len(reelSel) > 0 {
		var reelKids []string
		for _, sel := range reelSel {
			reelKids = append(reelKids, sel+" > *")
		}
		emitPrimitive(".fl-reel > *", reelKids, []string{
			"scroll-snap-align: start;",
			"flex: 0 0 auto;",
		})
	}

	emitPrimitive(".fl-frame", frameSel, []string{
		"aspect-ratio: var(--ratio);",
		"overflow: hidden;",
		"display: flex;",
		"justify-content: center;",
		"align-items: center;",
	})
	if len(frameSel) > 0 {
		var frameKids []string
		for _, sel := range frameSel {
			frameKids = append(frameKids, sel+" > img", sel+" > video")
		}
		emitPrimitive(".fl-frame > img, .fl-frame > video", frameKids, []string{
			"width: 100%;",
			"height: 100%;",
			"object-fit: cover;",
		})
	}

	// Exceptions
	emitPrimitive(".exc-fill", fillSel, []string{
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})

	emitPrimitive(".exc-scrolls", scrollsSel, []string{
		"overflow-y: auto;",
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})

	emitPrimitive(".exc-fixed", fixedSel, []string{
		"flex-shrink: 0;",
		"flex-grow: 0;",
	})

	emitPrimitive(".exc-flush", flushSel, []string{
		"margin: 0;",
		"border-radius: 0;",
	})

	emitPrimitive(".exc-clip", clipSel, []string{
		"overflow: hidden;",
	})

	sb.WriteString("}\n\n")

	// 3. Emit @layer widgets
	sb.WriteString("@layer widgets {\n")

	// Root rule
	rootDecls := s.RootRule.Decls()
	if len(rootDecls) > 0 {
		sb.WriteString(formatRule([]string{selectorOf(s.Name, "")}, rootDecls))
	}

	// Part rules (sorted)
	for _, p := range parts {
		partDecls := s.PartRules[widget.Part(p)].Decls()
		if len(partDecls) > 0 {
			sb.WriteString(formatRule([]string{selectorOf(s.Name, widget.Part(p))}, partDecls))
		}
	}

	sb.WriteString("}\n\n")

	// 4. Emit @layer states
	sb.WriteString("@layer states {\n")

	// Sort state rules: by state value, then part name
	type sortedState struct {
		key  stateKey
		rule Rule
	}
	var sortedStates []sortedState
	for k, r := range s.StateRules {
		sortedStates = append(sortedStates, sortedState{key: k, rule: r})
	}
	sort.Slice(sortedStates, func(i, j int) bool {
		if sortedStates[i].key.State != sortedStates[j].key.State {
			return sortedStates[i].key.State < sortedStates[j].key.State
		}
		return sortedStates[i].key.Part < sortedStates[j].key.Part
	})

	for _, ss := range sortedStates {
		decls := ss.rule.Decls()
		if len(decls) > 0 {
			attr := ss.key.State.Attr()
			sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.Name, ss.key.Part), attr.Key, attr.Value)
			sb.WriteString(formatRule([]string{sel}, decls))
		}
	}

	// Sort cue rules: by cue value, then part name
	type sortedCue struct {
		key  cueKey
		rule Rule
	}
	var sortedCues []sortedCue
	for k, r := range s.CueRules {
		sortedCues = append(sortedCues, sortedCue{key: k, rule: r})
	}
	sort.Slice(sortedCues, func(i, j int) bool {
		if sortedCues[i].key.Cue != sortedCues[j].key.Cue {
			return sortedCues[i].key.Cue < sortedCues[j].key.Cue
		}
		return sortedCues[i].key.Part < sortedCues[j].key.Part
	})

	for _, sc := range sortedCues {
		decls := sc.rule.Decls()
		if len(decls) > 0 {
			sel := selectorOf(s.Name, sc.key.Part) + cuePseudo(sc.key.Cue)
			sb.WriteString(formatRule([]string{sel}, decls))
		}
	}

	sb.WriteString("}\n")

	// Use css.Raw to produce our final, perfectly structured css.Stylesheet
	return css.NewStylesheet(css.Raw(sb.String()))
}
