//go:build !wasm

package style

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func (s *Sheet) Validate() []error {
	var errs []error

	for k := range s.stateRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, fmt.Errf("sheet %s: rule for undeclared part %q", string(s.widget.WidgetName()), string(k.part)))
			}
		}
	}
	for k := range s.cueRules {
		if k.part != "" {
			if _, exists := s.partRules[k.part]; !exists {
				errs = append(errs, fmt.Errf("sheet %s: rule for undeclared part %q", string(s.widget.WidgetName()), string(k.part)))
			}
		}
	}

	for p, pr := range s.partRules {
		if pr.hidden {
			continue
		}
		if pr.emitsNothing(s.widget.WidgetKind().Layer()) {
			errs = append(errs, fmt.Errf("sheet %s: part %q emits nothing", string(s.widget.WidgetName()), string(p)))
		}
	}

	checkVeil := func(p widget.Part, r rule) {
		if r.hasVeil && !r.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(p)))
		}
	}
	checkVeil("", s.rootRule)
	for p, pr := range s.partRules {
		checkVeil(p, pr)
	}

	// Every one of these sets `position`, and the emitted declarations are
	// sorted, so two of them on one rule means the alphabetically later keyword
	// silently wins. Anchor() alongside Docked()/Flyout() is the easy mistake:
	// both of those are already containing blocks, so the Anchor is redundant
	// as well as destructive.
	checkPosition := func(p widget.Part, r rule) {
		n := 0
		for _, on := range []bool{r.hasAnchor, r.hasDocked, r.hasOnEdge, r.hasFlyout, r.hasBackdrop, r.hasDrawer} {
			if on {
				n++
			}
		}
		if n > 1 {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Anchor/Docked/OnEdge/Flyout/Backdrop/Drawer all set position; use one", string(s.widget.WidgetName()), string(p)))
		}
	}
	// Both ends of a descendant rule have to exist, or it silently styles
	// nothing. CueWithinHover carries the same obligation.
	checkCueWithin := func(method string, k cueWithinKey) {
		if _, ok := s.partRules[k.container]; !ok && k.container != "" {
			errs = append(errs, fmt.Errf("sheet %s: %s container %q is not a declared part", string(s.widget.WidgetName()), method, string(k.container)))
		}
		if _, ok := s.partRules[k.part]; !ok && k.part != "" {
			errs = append(errs, fmt.Errf("sheet %s: %s part %q is not a declared part", string(s.widget.WidgetName()), method, string(k.part)))
		}
		if k.container == k.part {
			errs = append(errs, fmt.Errf("sheet %s: %s container and part are both %q; use Cue()", string(s.widget.WidgetName()), method, string(k.part)))
		}
	}
	for k := range s.cueWithin {
		checkCueWithin("CueWithin", k)
	}
	for k := range s.cueWithinHover {
		checkCueWithin("CueWithinHover", k)
	}

	// A device rule that only paints is invisible: OnlyOn hides the part by
	// default, and inside the query only a flow, CenterContent or the state rule
	// RevealedBy generates puts a `display` back on it.
	for k, dr := range s.deviceRules {
		pr, ok := s.partRules[k.part]
		if !ok || !pr.hidden || dr.hidden {
			continue
		}
		if !dr.hasFlow && !dr.centerContent && !dr.startContent && !dr.hasRevealed {
			errs = append(errs, fmt.Errf("sheet %s: part %q is OnlyOn but its device rule declares no flow, so it can never show", string(s.widget.WidgetName()), string(k.part)))
		}
	}

	checkPosition("", s.rootRule)
	for p, pr := range s.partRules {
		checkPosition(p, pr)
	}
	for k, dr := range s.deviceRules {
		checkPosition(k.part, dr)
	}
	for k, sr := range s.stateRules {
		if sr.hasVeil && !sr.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(k.part)))
		}
	}
	for k, cr := range s.cueRules {
		if cr.hasVeil && !cr.hasBackdrop {
			errs = append(errs, fmt.Errf("sheet %s: part %q: Veil() requires Backdrop()", string(s.widget.WidgetName()), string(k.part)))
		}
	}

	for k := range s.stateRules {
		if !s.widget.WidgetKind().Allows(k.state) {
			errs = append(errs, fmt.Errf("sheet %s: part %q: state %s is not meaningful for kind %s", string(s.widget.WidgetName()), string(k.part), k.state.String(), s.widget.WidgetKind().String()))
		}
	}

	checkInteractive := func(p widget.Part, r rule) {
		if r.interactive && (r.surface == Page || r.surface == Inactive) {
			errs = append(errs, fmt.Errf("sheet %s: part %q: surface %s has no interaction states", string(s.widget.WidgetName()), string(p), r.surface.String()))
		}
	}
	checkInteractive("", s.rootRule)
	for p, pr := range s.partRules {
		checkInteractive(p, pr)
	}

	hasDrawer := make(map[widget.Part]bool)
	hasRevealed := make(map[widget.Part]bool)
	checkPart := func(p widget.Part, r rule) {
		if r.hasDrawer {
			hasDrawer[p] = true
		}
		if r.hasRevealed {
			hasRevealed[p] = true
		}
	}
	checkPart("", s.rootRule)
	for p, pr := range s.partRules {
		checkPart(p, pr)
	}
	for p := range hasDrawer {
		if !hasRevealed[p] {
			if p == "" {
				errs = append(errs, fmt.Errf("sheet %s: root: Drawer() without RevealedBy(); the panel would be permanently visible", string(s.widget.WidgetName())))
			} else {
				errs = append(errs, fmt.Errf("sheet %s: part %q: Drawer() without RevealedBy(); the panel would be permanently visible", string(s.widget.WidgetName()), string(p)))
			}
		}
	}

	onlyOnDevices := make(map[widget.Part]css.Device)
	for dk := range s.deviceRules {
		if pr, exists := s.partRules[dk.part]; exists && pr.hidden {
			if prev, exists := onlyOnDevices[dk.part]; exists {
				errs = append(errs, fmt.Errf("sheet %s: part %q declared OnlyOn for both %s and %s", string(s.widget.WidgetName()), string(dk.part), prev.String(), dk.device.String()))
			} else {
				onlyOnDevices[dk.part] = dk.device
			}
		}
	}

	return errs
}
