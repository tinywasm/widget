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

// Anchor makes the element the positioning reference for a Flyout inside it.
// It is the trigger's container — a menu, a combobox — and emits nothing but
// position: relative.
func Anchor() Option {
	return func(r *rule) {
		r.hasAnchor = true
	}
}

// Flyout lifts the element out of the flow and hangs it under its Anchor, flush
// with the given inline edge, at the widget kind's stacking layer. Use it for a
// dropdown: left in the flow, an expanding menu pushes everything below it down
// and the list jumps under the pointer that opened it.
//
// The nearest Anchor() ancestor is what it hangs from. Without one it falls
// back to whatever ancestor happens to be positioned.
func Flyout(side Side) Option {
	return func(r *rule) {
		r.hasFlyout = true
		r.flyoutSide = side
	}
}

// Docked pins the element to the bottom corner of its Anchor, above the content
// and out of the flow, at the widget kind's stacking layer. It is the floating
// action button of a narrow screen: the list keeps the whole panel and the
// action stays reachable without a band of its own.
func Docked(side Side, gap Space) Option {
	return func(r *rule) {
		r.hasDocked = true
		r.dockedSide = side
		r.dockedGap = gap
	}
}

// Drawer anchors the element to one inline edge of the viewport, full height,
// at the widget kind's stacking layer. It is the slide-in panel of a mobile
// navigation; pair it with RevealedBy(widget.Open) to control visibility and
// with a sibling Backdrop(Viewport)+Veil() for the dimmed page behind it.
//
// Drawer sets the element's width. Do NOT also pass Width() — Validate rejects it.
func Drawer(side Side, size Size) Option {
	return func(r *rule) {
		r.hasDrawer = true
		r.drawerSide = side
		r.drawerSize = size
	}
}
