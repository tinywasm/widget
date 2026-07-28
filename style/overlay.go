//go:build !wasm

package style

// Scope dice contra qué se dimensiona un overlay.
type Scope uint8

const (
	// Parent cubre el ancestro posicionado más cercano (position: absolute).
	Parent Scope = iota
	// Viewport cubre toda la ventana (position: fixed).
	Viewport
)

// Backdrop saca el elemento del flujo y lo estira sobre todo su Scope. Es la capa
// que respalda a otra: un cazaclics o un velo. Queda por encima del contenido
// normal y por debajo de cualquier Above().
//
// Se llama Backdrop y no Overlay porque Overlay ya es el nivel máximo del enum
// Elevation (scale.go), que consumidores publicados usan como Raise(Overlay).
func Backdrop(s Scope) Opt {
	return func(r *Rule) {
		r.HasBackdrop = true
		r.BackdropScope = s
	}
}

// Above coloca el elemento por encima de cualquier Backdrop: el desplegable de un
// menú, el panel de un diálogo. No lo saca del flujo por sí solo.
func Above() Opt {
	return func(r *Rule) {
		r.Above = true
	}
}

// Scrim rellena el elemento con un velo translúcido sobre la superficie de la
// página. Solo tiene sentido junto a Backdrop.
func Scrim() Opt {
	return func(r *Rule) {
		r.Scrim = true
	}
}

// Hidden oculta el elemento. Es la regla base de un overlay que solo debe verse
// en cierto estado; se revierte con Shown() en una regla When().
func Hidden() Opt {
	return func(r *Rule) {
		r.Hidden = true
	}
}

// Shown revierte Hidden(). Va en una regla de estado, nunca en la base.
func Shown() Opt {
	return func(r *Rule) {
		r.Shown = true
	}
}
