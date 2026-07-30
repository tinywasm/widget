//go:build !wasm

package style

import (
	"github.com/tinywasm/css"
	"github.com/tinywasm/widget"
)

// Option is a visual option that configures a rule.
type Option func(*rule)

// rule contains all accumulated visual properties for an element.
type rule struct {
	hasFlow    bool
	flowType   flowType
	flowGap    Space
	flowRatio  SplitRatio
	flowAspect Aspect
	flowWidth  ColumnWidth
	flowSide   Side
	flowRail   RailWidth

	hasDrawer   bool
	drawerSide  Side
	drawerSize  Size

	hasSurface  bool
	surface     Surface
	interactive bool

	hasPad bool
	pad    Space

	hasRound bool
	round    Radius

	hasRaise bool
	raise    Elevation

	hasSize bool
	size    Size

	hasIcon bool
	icon    IconSize

	fill         bool
	scroll       bool
	keepSize     bool
	edgeToEdge   bool
	hideOverflow bool

	hasTextSize bool
	textSize    TextSize
	hasWeight   bool
	weight      Weight

	hasMotion bool
	motion    Motion

	hasBackdrop   bool
	backdropScope Scope
	hasVeil       bool
	revealedBy    widget.State
	hasRevealed   bool

	hidden bool
}

type stateKey struct {
	state widget.State
	part  widget.Part
}

type cueKey struct {
	cue  widget.Cue
	part widget.Part
}

type deviceKey struct {
	device css.Device
	part   widget.Part
}

// Sheet represents a scoped stylesheet for a widget.
type Sheet struct {
	widget      widget.Widget
	rootRule    rule
	partRules   map[widget.Part]rule
	stateRules  map[stateKey]rule
	cueRules    map[cueKey]rule
	deviceRules map[deviceKey]rule
}

// For opens the styling block for a widget.
func For(w widget.Widget) *Sheet {
	return &Sheet{
		widget:      w,
		partRules:   make(map[widget.Part]rule),
		stateRules:  make(map[stateKey]rule),
		cueRules:    make(map[cueKey]rule),
		deviceRules: make(map[deviceKey]rule),
	}
}

// Root defines the style for the root element of the widget.
func (s *Sheet) Root(opts ...Option) *Sheet {
	for _, opt := range opts {
		opt(&s.rootRule)
	}
	return s
}

// Part defines the style for an anatomical part of the widget.
func (s *Sheet) Part(p widget.Part, opts ...Option) *Sheet {
	r := s.partRules[p]
	for _, opt := range opts {
		opt(&r)
	}
	s.partRules[p] = r
	return s
}

// When defines the style for a part (or Root if p is "") when the widget has a specific state.
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Option) *Sheet {
	key := stateKey{state: st, part: p}
	r := s.stateRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.stateRules[key] = r
	return s
}

// Cue defines the style for a part (or Root if p is "") when the browser has a cue.
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Option) *Sheet {
	key := cueKey{cue: c, part: p}
	r := s.cueRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.cueRules[key] = r
	return s
}

// On defines the style for a part (or Root if p is "") only on the given viewport
// class. It is the single sanctioned way to vary a widget by device: the query
// strings live in tinywasm/css and are exhaustively tested there.
//
// Reach for a flow primitive first — Split, Grid and Sidebar already reflow on
// their own. Use On only when the ARRANGEMENT itself differs, e.g. a nav rail
// that becomes a drawer.
func (s *Sheet) On(d css.Device, p widget.Part, opts ...Option) *Sheet {
	key := deviceKey{device: d, part: p}
	r := s.deviceRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.deviceRules[key] = r
	return s
}

// OnlyOn declares a part that exists on one viewport class and nowhere else:
// it is display:none by default and takes the given options only on d.
//
// Use it for chrome that is genuinely device-specific — a hamburger button, a
// drawer's backdrop. If the element merely CHANGES between devices rather than
// disappearing, declare it with Part() and refine it with On().
func (s *Sheet) OnlyOn(d css.Device, p widget.Part, opts ...Option) *Sheet {
	if _, exists := s.partRules[p]; !exists {
		s.partRules[p] = rule{hidden: true}
	}
	key := deviceKey{device: d, part: p}
	r := s.deviceRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.deviceRules[key] = r
	return s
}
