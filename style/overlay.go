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

// Docked pins the element inside a corner, above the content and out of the
// flow, at the widget kind's stacking layer. Use it for a control that must not
// cost the content a band of its own: a floating action button, a row's
// overflow menu.
//
// Parent pins it to the corner of the nearest Anchor. Viewport pins it to the
// screen, so it stays put while the content behind it scrolls or swipes — and
// it disappears with the widget, because a fixed descendant of a display:none
// ancestor is not rendered either.
func Docked(scope Scope, edge Edge, side Side, gap Space) Option {
	return func(r *rule) {
		r.hasDocked = true
		r.dockedScope = scope
		r.dockedEdge = edge
		r.dockedSide = side
		r.dockedGap = gap
	}
}

// OnEdge centres the element ON one of its Anchor's edge lines — half outside
// the box, half inside — the way a fieldset legend rides the border it labels.
//
// block is the distance from the Anchor's border to the line being ridden: pass
// the Anchor's padding to ride the box that padding encloses, or SpaceNone to
// ride the Anchor's own border. inline is how far the chip is indented along
// that line. The straddle is exact because the chip's height is the shared
// --chip-height token, applied as half a chip-height of negative margin.
func OnEdge(edge Edge, side Side, block Space, inline Space) Option {
	return func(r *rule) {
		r.hasOnEdge = true
		r.onEdgeEdge = edge
		r.onEdgeSide = side
		r.onEdgeBlock = block
		r.onEdgeInline = inline
	}
}

// FloatingChrome declares that this element occupies a strip along its edge,
// and every Scroll() region DESCENDANT of it must reserve that strip — the
// contract between a floating action button and the scroll container behind
// it. It emits, on its own box:
//
//	--floating-bottom: calc(<size> + 2 * <gap>);
//
// (and the --floating-top counterpart for EdgeTop). The custom property is
// inherited, so the reservation crosses widget and repository boundaries:
// the host says "I occupy this band of my edge" and a Scroll() descendant —
// whatever widget it belongs to — pads itself by var(--floating-bottom, 0px)
// without either knowing the other's class name. No FloatingChrome means no
// declaration, and every scroller reserves nothing (the 0px default).
//
// This is the seam the badge-over-FAB overlap needed: the FAB's box is
// invisible to the badge's scrollHeight, so no padding computed from the
// badge's own box could reserve the space where the FAB really paints.
func FloatingChrome(edge Edge, size Size, gap Space) Option {
	return func(r *rule) {
		r.hasFloatingChrome = true
		r.floatingChromeEdge = edge
		r.floatingChromeSize = size
		r.floatingChromeGap = gap
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
