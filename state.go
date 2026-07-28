package widget

import "github.com/tinywasm/fmt"

// State is a state that the widget POSSESSES: written by Go, read by the stylesheet.
type State uint8

const (
	Selected State = iota
	Disabled
	Locked  // read-only, but fully legible
	Invalid
	Busy
	Open    // deployed / expanded
	Current // active navigation item
)

func (s State) String() string {
	switch s {
	case Selected:
		return "Selected"
	case Disabled:
		return "Disabled"
	case Locked:
		return "Locked"
	case Invalid:
		return "Invalid"
	case Busy:
		return "Busy"
	case Open:
		return "Open"
	case Current:
		return "Current"
	default:
		return "Unknown"
	}
}

// Attr returns the attribute that the DOM writes and upon which the stylesheet selects.
// Markup and CSS match by construction, not by convention.
func (s State) Attr() fmt.KeyValue {
	switch s {
	case Selected:
		return fmt.KeyValue{Key: "data-selected", Value: "true"}
	case Disabled:
		return fmt.KeyValue{Key: "data-disabled", Value: "true"}
	case Locked:
		return fmt.KeyValue{Key: "data-locked", Value: "true"}
	case Invalid:
		return fmt.KeyValue{Key: "data-invalid", Value: "true"}
	case Busy:
		return fmt.KeyValue{Key: "data-busy", Value: "true"}
	case Open:
		return fmt.KeyValue{Key: "data-open", Value: "true"}
	case Current:
		return fmt.KeyValue{Key: "data-current", Value: "true"}
	default:
		return fmt.KeyValue{}
	}
}

// Cue is a state that the BROWSER possesses. It is only styled;
// it cannot be written from Go — which is why it is a separate type and has no Attr().
type Cue uint8

const (
	Hover Cue = iota
	Focus
	Press
	Target
)
