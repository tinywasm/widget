//go:build !wasm

package style

import (
	"sort"

	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func (s *Sheet) Stylesheet() *css.Stylesheet {
	if errs := s.Validate(); len(errs) > 0 {
		var msgs []any
		msgs = append(msgs, "widget/style:")
		for _, err := range errs {
			msgs = append(msgs, err.Error())
		}
		panic(fmt.Err(msgs...))
	}

	sb := fmt.GetConv()
	defer sb.PutConv()

	sb.WriteString("@layer tokens, primitives, widgets, states;\n\n")

	var stackSel, rowSel, splitSel, gridSel, centerSel, fillCenteredSel, scrollRowSel, mediaBoxSel []string
	var coverSels []string

	type sidebarInfo struct {
		sel  string
		side Side
	}
	var sidebarInfos []sidebarInfo

	var fillSel, growSel, scrollSel, keepSizeSel, edgeToEdgeSel, hideOverflowSel []string

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
			case flowCover:
				coverSels = append(coverSels, sel)
			case flowSidebar:
				sidebarInfos = append(sidebarInfos, sidebarInfo{sel: sel, side: r.flowSide})
			}
		}
		if r.fill {
			fillSel = append(fillSel, sel)
		}
		if r.grow {
			growSel = append(growSel, sel)
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

	primitivesSB := fmt.GetConv()
	defer primitivesSB.PutConv()

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

	emitPrimitive(coverSels, []string{
		"height: 100dvh;",
		"display: flex;",
		"flex-direction: column;",
	})

	for _, si := range sidebarInfos {
		emitPrimitive([]string{si.sel}, []string{
			"display: flex;",
			"flex-wrap: wrap;",
			"gap: var(--gap);",
		})
		emitPrimitive([]string{sidebarRailSel(si.sel, si.side)}, []string{
			"flex-basis: var(--rail);",
			"flex-grow: 1;",
		})
		emitPrimitive([]string{sidebarContentSel(si.sel, si.side)}, []string{
			"flex-basis: 0;",
			"flex-grow: 999;",
			"min-width: 50%;",
		})
	}

	emitPrimitive(fillSel, []string{
		"height: 100%;",
		"min-height: 0;",
		"flex-grow: 1;",
	})
	emitPrimitive(growSel, []string{
		"flex-grow: 1;",
		"min-width: 0;",
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

	prims := primitivesSB.GetString(fmt.BuffOut)
	if len(prims) > 0 {
		sb.WriteString("@layer primitives {\n")
		sb.WriteString(prims)
		sb.WriteString("}\n\n")
	}

	widgetsSB := fmt.GetConv()
	defer widgetsSB.PutConv()

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

	wids := widgetsSB.GetString(fmt.BuffOut)
	if len(wids) > 0 {
		sb.WriteString("@layer widgets {\n")
		sb.WriteString(wids)
		sb.WriteString("}\n\n")
	}

	statesSB := fmt.GetConv()
	defer statesSB.PutConv()

	stateDecls := make(map[stateKey][]string)
	for k, sr := range s.stateRules {
		stateDecls[k] = sr.Decls(s.widget.WidgetKind().Layer())
	}

	if s.rootRule.hasRevealed {
		sk := stateKey{state: s.rootRule.revealedBy, part: ""}
		stateDecls[sk] = append(stateDecls[sk], "display: "+displayFor(s.rootRule.flowType)+";")
	}
	for _, p := range parts {
		pr := s.partRules[p]
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

	states := statesSB.GetString(fmt.BuffOut)
	if len(states) > 0 {
		sb.WriteString("@layer states {\n")
		sb.WriteString(states)
		sb.WriteString("}\n\n")
	}

	var deviceOrder []css.Device
	for dk := range s.deviceRules {
		deviceOrder = append(deviceOrder, dk.device)
	}
	sort.Slice(deviceOrder, func(i, j int) bool {
		return deviceOrder[i] < deviceOrder[j]
	})
	var deduped []css.Device
	for _, d := range deviceOrder {
		if len(deduped) == 0 || deduped[len(deduped)-1] != d {
			deduped = append(deduped, d)
		}
	}

	for _, d := range deduped {
		var deviceParts []deviceKey
		for dk := range s.deviceRules {
			if dk.device == d {
				deviceParts = append(deviceParts, dk)
			}
		}
		sort.Slice(deviceParts, func(i, j int) bool {
			return deviceParts[i].part < deviceParts[j].part
		})

		devSB := fmt.GetConv()

		for _, dk := range deviceParts {
			r := s.deviceRules[dk]
			sel := selectorOf(s.widget.WidgetName(), dk.part)

			devWidSB := fmt.GetConv()
			if r.hasFlow {
				switch r.flowType {
				case flowStack:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-direction: column;", "min-height: 0;"}))
					devWidSB.WriteString(formatRule([]string{sel + " > * + *"}, []string{"margin-block-start: var(--gap);"}))
				case flowRow:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);", "align-items: center;"}))
				case flowSplit:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, []string{"flex-grow: 1;", "flex-basis: calc((40rem - 100%) * 999);"}))
					devWidSB.WriteString(formatRule([]string{sel + " > :first-child"}, []string{"flex-grow: var(--ratio);"}))
				case flowGrid:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: grid;", "gap: var(--gap);", "grid-template-columns: repeat(auto-fit, minmax(min(var(--track), 100%), 1fr));"}))
				case flowCenter:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"margin-inline: auto;", "max-width: var(--max-width);", "width: 100%;"}))
				case flowFillCentered:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: grid;", "place-items: center;", "min-height: 100%;", "width: 100%;"}))
				case flowScrollRow:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "gap: var(--gap);", "overflow-x: auto;", "scroll-snap-type: x mandatory;"}))
					devWidSB.WriteString(formatRule([]string{sel + " > *"}, []string{"scroll-snap-align: start;", "flex: 0 0 auto;"}))
				case flowMediaBox:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"aspect-ratio: var(--ratio);", "overflow: hidden;", "display: flex;", "justify-content: center;", "align-items: center;"}))
					devWidSB.WriteString(formatRule([]string{sel + " > img", sel + " > video"}, []string{"width: 100%;", "height: 100%;", "object-fit: cover;"}))
				case flowCover:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"height: 100dvh;", "display: flex;", "flex-direction: column;"}))
				case flowSidebar:
					devWidSB.WriteString(formatRule([]string{sel}, []string{"display: flex;", "flex-wrap: wrap;", "gap: var(--gap);"}))
					devWidSB.WriteString(formatRule([]string{sidebarRailSel(sel, r.flowSide)}, []string{"flex-basis: var(--rail);", "flex-grow: 1;"}))
					devWidSB.WriteString(formatRule([]string{sidebarContentSel(sel, r.flowSide)}, []string{"flex-basis: 0;", "flex-grow: 999;", "min-width: 50%;"}))
				}
			}

			wd := r.Decls(s.widget.WidgetKind().Layer())
			if len(wd) > 0 {
				devWidSB.WriteString(formatRule([]string{sel}, wd))
			}

			devWid := devWidSB.GetString(fmt.BuffOut)
			devWidSB.PutConv()
			if len(devWid) > 0 {
				devSB.WriteString("@layer widgets {\n")
				devSB.WriteString(devWid)
				devSB.WriteString("}\n")
			}

			if r.hasRevealed {
				sk := stateKey{state: r.revealedBy, part: dk.part}
				attr := sk.state.Attr()
				stateSel := fmt.Sprintf("%s[%s=\"%s\"]", sel, attr.Key, attr.Value)
				devSB.WriteString(formatRule([]string{stateSel}, []string{"display: " + displayFor(r.flowType) + ";"}))
			}
		}

		devRules := devSB.GetString(fmt.BuffOut)
		devSB.PutConv()
		if len(devRules) > 0 {
			sb.WriteString("@media " + d.Query() + " {\n")
			sb.WriteString(devRules)
			sb.WriteString("}\n")
		}
	}

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

	return css.NewStylesheet(css.Raw(sb.GetString(fmt.BuffOut)))
}
