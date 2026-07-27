package widget

import "github.com/tinywasm/fmt"

// State es un estado que POSEE el widget: lo escribe Go, lo lee la hoja de estilos.
type State uint8

const (
	Selected State = iota
	Disabled
	Locked   // solo lectura, pero plenamente legible
	Invalid
	Busy
	Open     // desplegado / expandido
	Current  // ítem de navegación activo
)

// Attr devuelve el atributo que el DOM escribe y sobre el que la hoja selecciona.
// Markup y CSS coinciden por construcción, no por convención.
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

// Cue es un estado que posee el NAVEGADOR. Solo se estiliza; no se puede escribir
// desde Go — por eso es un tipo distinto y no tiene Attr().
type Cue uint8

const (
	Hover Cue = iota
	Focus
	Press
	Target
)
