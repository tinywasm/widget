//go:build !wasm

package style

import "github.com/tinywasm/widget"

// Scope says what an overlay dimensions against.
type Scope uint8

const (
	// Parent covers the nearest positioned ancestor (position: absolute).
	Parent Scope = iota
	// Viewport covers the entire window (position: fixed).
	Viewport
)

// Backdrop removes the element from the normal flow and stretches it over its Scope.
func Backdrop(s Scope) Option {
	return func(r *rule) {
		r.hasBackdrop = true
		r.backdropScope = s
	}
}

// Veil fills the element with a translucent wash overlaying the surface.
// Only makes sense alongside Backdrop.
func Veil() Option {
	return func(r *rule) {
		r.hasVeil = true
	}
}

// RevealedBy binds hiding/showing the element to a widget State.
func RevealedBy(st widget.State) Option {
	return func(r *rule) {
		r.hasRevealed = true
		r.revealedBy = st
	}
}
