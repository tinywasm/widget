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
	flowDetail Size
	flowMotion Motion
	flowCols   int

	hasDrawer  bool
	drawerSide Side
	drawerSize Size

	hasEdgeStrip   bool
	edgeStripScope Scope
	edgeStripSide  Side

	hasMeter       bool
	meterThickness Space
	centerSelf     bool

	centerContent bool
	startContent  bool
	controlBox    bool
	logoBox       bool
	chipBox       bool
	capitalize    bool

	hasGlyph bool
	glyph    Surface

	hasDocked   bool
	dockedScope Scope
	dockedEdge  Edge
	dockedSide  Side
	dockedGap   Space

	hasOnEdge    bool
	onEdgeEdge   Edge
	onEdgeSide   Side
	onEdgeBlock  Space
	onEdgeInline Space

	hasFloatingChrome  bool
	floatingChromeEdge Edge
	floatingChromeSize IconSize
	floatingChromeGap  Space

	hasAnchor  bool
	foreground bool
	hasFlyout  bool
	flyoutSide Side

	hasDivider  bool
	dividerSide Side

	hasSurface  bool
	surface     Surface
	interactive bool

	// overlay marca una regla de ESTADO (When/Cue/CueWithin). Un estado se pinta
	// encima de la caja base: no puede cambiar su tamaño, porque el elemento
	// crecería justo cuando el puntero está encima y perdería el propio hover que
	// lo activó. Es lo que hace que As() emita un anillo de box-shadow (ver
	// boxShadowDecls) en vez de border.
	overlay bool

	hasPad bool
	pad    Space

	hasPadEdge   bool
	padEdge      Edge
	padEdgeSpace Space

	hasPadInline bool
	padInline    Space

	hasRound bool
	round    Radius

	hasRaise bool
	raise    Elevation

	hasSize bool
	size    Size

	hasIcon bool
	icon    IconSize

	fill         bool
	grow         bool
	pushEnd      bool
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
	shown  bool
}

type stateKey struct {
	state widget.State
	part  widget.Part
}

type cueKey struct {
	cue  widget.Cue
	part widget.Part
}

// cueWithinKey addresses a part through an ancestor's cue. Every other rule in
// this package is one selector on one element; this is the single exception,
// and it exists because a container that reveals its own children on hover — a
// nav rail that expands — has no other expression.
type cueWithinKey struct {
	cue       widget.Cue
	container widget.Part
	part      widget.Part
}

// stateWithinKey addresses a part through an ancestor's STATE — the written
// counterpart of cueWithinKey. It exists because a state is written onto the
// element that owns it, which is not always the element it should repaint: a
// field's read-only gate belongs to the field, but what has to look read-only
// is the control inside it.
type stateWithinKey struct {
	state     widget.State
	container widget.Part
	part      widget.Part
}

type deviceKey struct {
	device css.Device
	part   widget.Part
}

// Sheet represents a scoped stylesheet for a widget.
type Sheet struct {
	widget         widget.Widget
	rootRule       rule
	partRules      map[widget.Part]rule
	stateRules     map[stateKey]rule
	stateWithin    map[stateWithinKey]rule
	cueRules       map[cueKey]rule
	cueWithin      map[cueWithinKey]rule
	cueWithinHover map[cueWithinKey]rule
	deviceRules    map[deviceKey]rule

	// within records the part tree: which part renders inside which. The
	// sheet needs it to reason about positioning (who is whose containing
	// block), because a Flyout's inset resolves against the nearest
	// POSITIONED ancestor, which is not always the Anchor the author meant.
	within map[widget.Part]widget.Part
}

// For opens the styling block for a widget.
func For(w widget.Widget) *Sheet {
	return &Sheet{
		widget:         w,
		partRules:      make(map[widget.Part]rule),
		stateRules:     make(map[stateKey]rule),
		stateWithin:    make(map[stateWithinKey]rule),
		cueRules:       make(map[cueKey]rule),
		cueWithin:      make(map[cueWithinKey]rule),
		cueWithinHover: make(map[cueWithinKey]rule),
		deviceRules:    make(map[deviceKey]rule),
		within:         make(map[widget.Part]widget.Part),
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

// Within declares that part renders INSIDE container, and applies the options
// to part exactly as Part() would; what it adds is the containment relation,
// which the sheet needs to reason about positioning — who is whose containing
// block. It reads like the DOM: Within("menu", "options", Flyout(...)) is
// "options, inside menu".
//
// Part() remains the normal declaration. Within() is only needed where
// containment changes the result: a Flyout that hangs from an Anchor while a
// positioned part sits between them. When it matters, the sheet rejects the
// composition until the nesting is declared — see Validate().
func (s *Sheet) Within(container, p widget.Part, opts ...Option) *Sheet {
	r := s.partRules[p]
	for _, opt := range opts {
		opt(&r)
	}
	s.partRules[p] = r
	s.within[p] = container
	return s
}

// When defines the style for a part (or Root if p is "") when the widget has a specific state.
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Option) *Sheet {
	key := stateKey{state: st, part: p}
	r := s.stateRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.stateRules[key] = r
	return s
}

// WhenWithin styles a part while an ANCESTOR part carries a written state —
// `.n__container[data-x="true"] .n__part`. It is the State counterpart of
// CueWithin, and like CueWithin it is the exception, not the habit: reach for
// When() first.
//
// It exists because dom writes a state onto the element that OWNS it, which is
// not always the element that should change. A form field's read-only gate is
// written on the field, but what must stop looking editable is the control
// inside it — When(Locked, PartInput) would emit
// `.n__input[data-locked="true"]` and match nothing, since the attribute is on
// the wrapper. Pass "" as container to hang the rule off the widget root.
func (s *Sheet) WhenWithin(st widget.State, container, p widget.Part, opts ...Option) *Sheet {
	key := stateWithinKey{state: st, container: container, part: p}
	r := s.stateWithin[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.stateWithin[key] = r
	return s
}

// Cue defines the style for a part (or Root if p is "") when the browser has a cue.
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Option) *Sheet {
	key := cueKey{cue: c, part: p}
	r := s.cueRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueRules[key] = r
	return s
}

// CueWithin styles a part while an ANCESTOR part carries a browser cue —
// `.n__container:hover .n__part`. Reach for Cue() first; this is only for the
// case where the trigger and the thing that reacts are different elements, such
// as a rail that shows its labels while the pointer is over it.
func (s *Sheet) CueWithin(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet {
	key := cueWithinKey{cue: c, container: container, part: p}
	r := s.cueWithin[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueWithin[key] = r
	return s
}

// CueWithinHover is CueWithin gated on the fine-pointer capability: the same
// descendant selector, emitted inside `@media (hover: hover)`. A touch tap
// fires `:hover` and synthetic mouse events, so a hover reveal that is not
// scoped this way misfires on a phone — the exact reason this variant exists.
func (s *Sheet) CueWithinHover(c widget.Cue, container, p widget.Part, opts ...Option) *Sheet {
	key := cueWithinKey{cue: c, container: container, part: p}
	r := s.cueWithinHover[key]
	for _, opt := range opts {
		opt(&r)
	}
	r.overlay = true
	s.cueWithinHover[key] = r
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
