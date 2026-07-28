//go:build !wasm

package style

import "github.com/tinywasm/widget"

// Opt es una opción visual que configura una regla.
type Opt func(*Rule)

// Rule contiene todas las propiedades visuales acumuladas para un elemento (Root o Part).
type Rule struct {
	HasFlow   bool
	FlowType  string // "stack", "row", "split", "grid", "center", "cover", "reel", "frame"
	FlowGap   Space
	FlowRatio Ratio
	FlowTrack Track

	HasSurface bool
	Surface    Surface

	HasPad bool
	Pad    Space

	HasRound bool
	Round    Radius

	HasRaise bool
	Raise    Elevation

	HasSize bool
	Size    Size

	Fill    bool
	Scrolls bool
	Fixed   bool
	Flush   bool
	Clip    bool

	HasTextSize bool
	TextSize    TextSize
	HasWeight   bool
	Weight      Weight

	HasOverlay   bool
	OverlayScope Scope
	Above        bool
	Scrim        bool
	Hidden       bool
	Shown        bool
}

type stateKey struct {
	State widget.State
	Part  widget.Part
}

type cueKey struct {
	Cue  widget.Cue
	Part widget.Part
}

// Sheet representa el bloque de estilo scoped a un widget.
type Sheet struct {
	Name       widget.Name
	RootRule   Rule
	PartRules  map[widget.Part]Rule
	StateRules map[stateKey]Rule
	CueRules   map[cueKey]Rule
}

// Of abre el bloque de estilo de un widget. Todas sus reglas quedan scoped a n.
func Of(n widget.Name) *Sheet {
	return &Sheet{
		Name:       n,
		PartRules:  make(map[widget.Part]Rule),
		StateRules: make(map[stateKey]Rule),
		CueRules:   make(map[cueKey]Rule),
	}
}

// Root define el estilo de la raíz del widget.
func (s *Sheet) Root(opts ...Opt) *Sheet {
	for _, opt := range opts {
		opt(&s.RootRule)
	}
	return s
}

// Part define el estilo de una parte anatómica del widget.
func (s *Sheet) Part(p widget.Part, opts ...Opt) *Sheet {
	r := s.PartRules[p]
	for _, opt := range opts {
		opt(&r)
	}
	s.PartRules[p] = r
	return s
}

// When define el estilo de una parte (o Root si p es "") cuando el widget posee un estado.
func (s *Sheet) When(st widget.State, p widget.Part, opts ...Opt) *Sheet {
	key := stateKey{State: st, Part: p}
	r := s.StateRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.StateRules[key] = r
	return s
}

// Cue define el estilo de una parte (o Root si p es "") cuando el navegador posee un cue (hover, focus, etc.).
func (s *Sheet) Cue(c widget.Cue, p widget.Part, opts ...Opt) *Sheet {
	key := cueKey{Cue: c, Part: p}
	r := s.CueRules[key]
	for _, opt := range opts {
		opt(&r)
	}
	s.CueRules[key] = r
	return s
}

// Styler es la capacidad "este widget tiene aspecto". La asevera el recolector SSR.
type Styler interface {
	widget.Widget
	Style() *Sheet
}
