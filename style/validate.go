//go:build !wasm

package style

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/widget"
)

func (s *Sheet) Validate() []error {
	var errs []error
	errs = s.validateParts(errs)
	errs = s.validateComposition(errs)
	errs = s.validateStates(errs)
	return errs
}

// validateParts is the declaration-integrity half of Validate(): a rule
// must reference only DECLARED parts, a declared part must actually emit
// something, and the base surface rules must satisfy their Backdrop()
// pairing.
func (s *Sheet) validateParts(errs []error) []error {
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
	return errs
}
