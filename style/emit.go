//go:build !wasm

package style

import (
	goFmt "fmt"
	"sort"
	"strings"

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

func splitRatioValue(r SplitRatio) string {
	switch r {
	case SplitHalf:
		return "1fr"
	case SplitTwoThirds:
		return "2fr"
	case SplitThreeQuarters:
		return "3fr"
	default:
		return "1fr"
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
	case flowStack, flowRow, flowScrollRow, flowMediaBox:
		return "flex"
	case flowSplit, flowGrid, flowFillCentered:
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

// rule.Decls formats the declarations of a rule
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

	return decls
}

// emitsNothing reports whether the rule contributes no CSS at all. Decls covers
// the widgets layer only; flows and exceptions are emitted into the primitives
// layer by collect, so both have to be consulted to answer the question.
func (r rule) emitsNothing(layer widget.Layer) bool {
	if len(r.Decls(layer)) > 0 {
		return false
	}
	return !r.hasFlow && !r.fill && !r.scroll && !r.keepSize && !r.edgeToEdge && !r.hideOverflow
}

// Formats a CSS rule block deterministically.
func formatRule(selectors []string, decls []string) string {
	if len(selectors) == 0 || len(decls) == 0 {
		return ""
	}
	sort.Strings(selectors)
	sort.Strings(decls)
	var sb strings.Builder
	sb.WriteString(strings.Join(selectors, ", ") + " {\n")
	for _, d := range decls {
		sb.WriteString("  " + d + "\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func familyBase(s Surface) css.Token {
	switch s {
	case Panel, Inset, Highlight:
		return css.ColorSurface
	case Primary:
		return css.ColorPrimary
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

// Stylesheet generates a css.Stylesheet for the Sheet.
func (s *Sheet) Stylesheet() *css.Stylesheet {
	if errs := s.Validate(); len(errs) > 0 {
		var msgs []any
		msgs = append(msgs, "widget/style:")
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		panic(fmt.Err(msgs...))
	}

	var sb strings.Builder

	// Fixed layer definition
	sb.WriteString("@layer tokens, primitives, widgets, states;\n\n")

	// 1. Gather selectors for primitives
	var stackSel, rowSel, splitSel, gridSel, centerSel, fillCenteredSel, scrollRowSel, mediaBoxSel []string
	var fillSel, scrollSel, keepSizeSel, edgeToEdgeSel, hideOverflowSel []string

	collect := func(r rule, sel string) {
		if r.hasFlow {
			switch r.flowType {
			case flowStack:
				stackSel = append(stackSel, sel)
			case flowRow:
				rowSel = append(rowSel, sel)
			case flowSplit:
				splitSel = append(splitSel, sel)
			case flowGrid:
				gridSel = append(gridSel, sel)
			case flowCenter:
				centerSel = append(centerSel, sel)
			case flowFillCentered:
				fillCenteredSel = append(fillCenteredSel, sel)
			case flowScrollRow:
				scrollRowSel = append(scrollRowSel, sel)
			case flowMediaBox:
				mediaBoxSel = append(mediaBoxSel, sel)
			}
		}
		if r.fill {
			fillSel = append(fillSel, sel)
		}
		if r.scroll {
			scrollSel = append(scrollSel, sel)
		}
		if r.keepSize {
			keepSizeSel = append(keepSizeSel, sel)
		}
		if r.edgeToEdge {
			edgeToEdgeSel = append(edgeToEdgeSel, sel)
		}
		if r.hideOverflow {
			hideOverflowSel = append(hideOverflowSel, sel)
		}
	}

	collect(s.rootRule, selectorOf(s.widget.WidgetName(), ""))

	var parts []widget.Part
	for p := range s.partRules {
		parts = append(parts, p)
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i] < parts[j]
	})

	for _, p := range parts {
		collect(s.partRules[p], selectorOf(s.widget.WidgetName(), p))
	}

	// 2. Emit @layer primitives (omitted entirely when empty)
	var primitivesSB strings.Builder

	emitPrimitive := func(extraSel []string, decls []string) {
		if len(extraSel) > 0 {
			primitivesSB.WriteString(formatRule(extraSel, decls))
		}
	}

	emitPrimitive(stackSel, []string{
		"display: flex;",
		"flex-direction: column;",
		"min-height: 0;",
	})
	if len(stackSel) > 0 {
		var stackKids []string
		for _, sel := range stackSel {
			stackKids = append(stackKids, sel+" > * + *")
		}
		emitPrimitive(stackKids, []string{
			"margin-block-start: var(--gap);",
		})
	}

	emitPrimitive(rowSel, []string{
		"display: flex;",
		"flex-wrap: wrap;",
		"gap: var(--gap);",
		"align-items: center;",
	})

	emitPrimitive(splitSel, []string{
		"display: flex;",
		"flex-wrap: wrap;",
		"gap: var(--gap);",
	})
	if len(splitSel) > 0 {
		var splitKids []string
		var splitFirst []string
		for _, sel := range splitSel {
			splitKids = append(splitKids, sel+" > *")
			splitFirst = append(splitFirst, sel+" > :first-child")
		}
		emitPrimitive(splitKids, []string{
			"flex-grow: 1;",
			"flex-basis: calc((40rem - 100%) * 999);",
		})
		emitPrimitive(splitFirst, []string{
			"flex-grow: var(--ratio);",
		})
	}

	emitPrimitive(gridSel, []string{
		"display: grid;",
		"gap: var(--gap);",
		"grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));",
	})

	emitPrimitive(centerSel, []string{
		"margin-inline: auto;",
		"max-width: var(--max-width);",
		"width: 100%;",
	})

	emitPrimitive(fillCenteredSel, []string{
		"display: grid;",
		"place-items: center;",
		"min-height: 100%;",
		"width: 100%;",
	})

	emitPrimitive(scrollRowSel, []string{
		"display: flex;",
		"gap: var(--gap);",
		"overflow-x: auto;",
		"scroll-snap-type: x mandatory;",
	})
	if len(scrollRowSel) > 0 {
		var scrollRowKids []string
		for _, sel := range scrollRowSel {
			scrollRowKids = append(scrollRowKids, sel+" > *")
		}
		emitPrimitive(scrollRowKids, []string{
			"scroll-snap-align: start;",
			"flex: 0 0 auto;",
		})
	}

	emitPrimitive(mediaBoxSel, []string{
		"aspect-ratio: var(--ratio);",
		"overflow: hidden;",
		"display: flex;",
		"justify-content: center;",
		"align-items: center;",
	})
	if len(mediaBoxSel) > 0 {
		var mediaBoxKids []string
		for _, sel := range mediaBoxSel {
			mediaBoxKids = append(mediaBoxKids, sel+" > img", sel+" > video")
		}
		emitPrimitive(mediaBoxKids, []string{
			"width: 100%;",
			"height: 100%;",
			"object-fit: cover;",
		})
	}

	// Exceptions
	emitPrimitive(fillSel, []string{
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})
	emitPrimitive(scrollSel, []string{
		"overflow-y: auto;",
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})
	emitPrimitive(keepSizeSel, []string{
		"flex-shrink: 0;",
		"flex-grow: 0;",
	})
	emitPrimitive(edgeToEdgeSel, []string{
		"margin: 0;",
		"border-radius: 0;",
	})
	emitPrimitive(hideOverflowSel, []string{
		"overflow: hidden;",
	})

	if primitivesSB.Len() > 0 {
		sb.WriteString("@layer primitives {\n" + primitivesSB.String() + "}\n\n")
	}

	// 3. Emit @layer widgets (omitted entirely when empty)
	var widgetsSB strings.Builder

	rootDecls := s.rootRule.Decls(s.widget.WidgetKind().Layer())
	if len(rootDecls) > 0 {
		widgetsSB.WriteString(formatRule([]string{selectorOf(s.widget.WidgetName(), "")}, rootDecls))
	}

	for _, p := range parts {
		partDecls := s.partRules[p].Decls(s.widget.WidgetKind().Layer())
		if len(partDecls) > 0 {
			widgetsSB.WriteString(formatRule([]string{selectorOf(s.widget.WidgetName(), p)}, partDecls))
		}
	}

	if widgetsSB.Len() > 0 {
		sb.WriteString("@layer widgets {\n" + widgetsSB.String() + "}\n\n")
	}

	// 4. Emit @layer states (omitted entirely when empty)
	var statesSB strings.Builder

	stateDecls := make(map[stateKey][]string)
	for k, sr := range s.stateRules {
		stateDecls[k] = sr.Decls(s.widget.WidgetKind().Layer())
	}

	// Inject Resolved display logic for RevealedBy
	if s.rootRule.hasRevealed {
		sk := stateKey{state: s.rootRule.revealedBy, part: ""}
		stateDecls[sk] = append(stateDecls[sk], "display: "+displayFor(s.rootRule.flowType)+";")
	}
	for p, pr := range s.partRules {
		if pr.hasRevealed {
			sk := stateKey{state: pr.revealedBy, part: p}
			stateDecls[sk] = append(stateDecls[sk], "display: "+displayFor(pr.flowType)+";")
		}
	}

	type sortedState struct {
		key   stateKey
		decls []string
	}
	var sortedStates []sortedState
	for k, decls := range stateDecls {
		if len(decls) > 0 {
			sortedStates = append(sortedStates, sortedState{key: k, decls: decls})
		}
	}
	sort.Slice(sortedStates, func(i, j int) bool {
		if sortedStates[i].key.state != sortedStates[j].key.state {
			return sortedStates[i].key.state < sortedStates[j].key.state
		}
		return sortedStates[i].key.part < sortedStates[j].key.part
	})

	for _, ss := range sortedStates {
		attr := ss.key.state.Attr()
		sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), ss.key.part), attr.Key, attr.Value)
		statesSB.WriteString(formatRule([]string{sel}, ss.decls))
	}

	// Cue Rules and Interactive Cues
	cueDecls := make(map[cueKey][]string)
	for k, cr := range s.cueRules {
		cueDecls[k] = cr.Decls(s.widget.WidgetKind().Layer())
	}

	addInteractive := func(p widget.Part, r rule) {
		if r.interactive {
			base := familyBase(r.surface)
			if base.Name != "" {
				kHover := cueKey{cue: widget.Hover, part: p}
				cueDecls[kHover] = append(cueDecls[kHover], "background-color: "+css.Hover(base)+";")

				kFocus := cueKey{cue: widget.Focus, part: p}
				cueDecls[kFocus] = append(cueDecls[kFocus], "background-color: "+css.Focus(base)+";")

				kPress := cueKey{cue: widget.Press, part: p}
				cueDecls[kPress] = append(cueDecls[kPress], "background-color: "+css.Press(base)+";")
			}
		}
	}

	addInteractive("", s.rootRule)
	for _, p := range parts {
		addInteractive(p, s.partRules[p])
	}

	type sortedCue struct {
		key   cueKey
		decls []string
	}
	var sortedCues []sortedCue
	for k, decls := range cueDecls {
		if len(decls) > 0 {
			sortedCues = append(sortedCues, sortedCue{key: k, decls: decls})
		}
	}
	sort.Slice(sortedCues, func(i, j int) bool {
		if sortedCues[i].key.cue != sortedCues[j].key.cue {
			return sortedCues[i].key.cue < sortedCues[j].key.cue
		}
		return sortedCues[i].key.part < sortedCues[j].key.part
	})

	for _, sc := range sortedCues {
		sel := selectorOf(s.widget.WidgetName(), sc.key.part) + cuePseudo(sc.key.cue)
		statesSB.WriteString(formatRule([]string{sel}, sc.decls))
	}

	if statesSB.Len() > 0 {
		sb.WriteString("@layer states {\n" + statesSB.String() + "}\n\n")
	}

	// Prefers reduced motion (only if any rule carries Animate)
	var motionSel []string
	if s.rootRule.hasMotion {
		motionSel = append(motionSel, selectorOf(s.widget.WidgetName(), ""))
	}
	for _, p := range parts {
		if s.partRules[p].hasMotion {
			motionSel = append(motionSel, selectorOf(s.widget.WidgetName(), p))
		}
	}
	for k, sr := range s.stateRules {
		if sr.hasMotion {
			attr := k.state.Attr()
			sel := fmt.Sprintf("%s[%s=\"%s\"]", selectorOf(s.widget.WidgetName(), k.part), attr.Key, attr.Value)
			motionSel = append(motionSel, sel)
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasMotion {
			sel := selectorOf(s.widget.WidgetName(), k.part) + cuePseudo(k.cue)
			motionSel = append(motionSel, sel)
		}
	}

	if len(motionSel) > 0 {
		sort.Strings(motionSel)
		sb.WriteString("@media (prefers-reduced-motion: reduce) {\n")
		sb.WriteString(formatRule(motionSel, []string{"transition: none;"}))
		sb.WriteString("}\n")
	}

	// Ensure there are no empty trailing lines or incorrect double spaces
	return css.NewStylesheet(css.Raw(sb.String()))
}

// Validate validates the Sheet according to SPECS.md §6.1.
func (s *Sheet) Validate() []error {
	var errs []error

	// 1. When/Cue names a part never declared with Part() (excluding empty part, which is root)
	for k := range s.stateRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, goFmt.Errorf("sheet %s: rule for undeclared part %q", s.widget.WidgetName(), k.part))
			}
		}
	}
	for k := range s.cueRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, goFmt.Errorf("sheet %s: rule for undeclared part %q", s.widget.WidgetName(), k.part))
			}
		}
	}

	// 2. A declared part produces no declarations
	for p, pr := range s.partRules {
		if pr.emitsNothing(s.widget.WidgetKind().Layer()) {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q emits nothing", s.widget.WidgetName(), p))
		}
	}

	// 3. Veil() without Backdrop() on the same rule
	checkVeil := func(p widget.Part, r rule) {
		if r.hasVeil && !r.hasBackdrop {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q: Veil() requires Backdrop()", s.widget.WidgetName(), p))
		}
	}
	checkVeil("", s.rootRule)
	for p, pr := range s.partRules {
		checkVeil(p, pr)
	}
	for k, sr := range s.stateRules {
		if sr.hasVeil && !sr.hasBackdrop {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q: Veil() requires Backdrop()", s.widget.WidgetName(), k.part))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasVeil && !cr.hasBackdrop {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q: Veil() requires Backdrop()", s.widget.WidgetName(), k.part))
		}
	}

	// 4. When uses a state Kind.Allows rejects
	for k := range s.stateRules {
		if !s.widget.WidgetKind().Allows(k.state) {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q: state %s is not meaningful for kind %s", s.widget.WidgetName(), k.part, k.state, s.widget.WidgetKind()))
		}
	}

	// 5. Interactive() on Page or Inactive
	checkInteractive := func(p widget.Part, r rule) {
		if r.interactive && (r.surface == Page || r.surface == Inactive) {
			errs = append(errs, goFmt.Errorf("sheet %s: part %q: surface %s has no interaction states", s.widget.WidgetName(), p, r.surface))
		}
	}
	checkInteractive("", s.rootRule)
	for p, pr := range s.partRules {
		checkInteractive(p, pr)
	}

	return errs
}

// Parts returns the declared parts, sorted.
func (s *Sheet) Parts() []widget.Part {
	var parts []widget.Part
	for p := range s.partRules {
		parts = append(parts, p)
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i] < parts[j]
	})
	return parts
}
