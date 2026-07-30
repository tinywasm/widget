//go:build !wasm

package style

// Fill takes up the entire available height.
func Fill() Option {
	return func(r *rule) {
		r.fill = true
	}
}

// Scroll overflows internally instead of growing. Implies Fill().
func Scroll() Option {
	return func(r *rule) {
		r.scroll = true
		r.fill = true
	}
}

// Grow takes the free space along the inline axis and nothing else. It is the
// Row counterpart of Fill(): Fill() also claims `height: 100%`, which inside a
// Row resolves against the row and stretches the part into a full-height block.
// Use Grow() for the item in a Row that should push its siblings to the
// trailing edge.
func Grow() Option {
	return func(r *rule) {
		r.grow = true
	}
}

// PushEnd sends the part to the trailing edge of its line. It is the companion
// of Grow(): Grow() absorbs the free space so the items after it are pushed
// out, PushEnd() moves the free space in front of a single item — the only way
// to keep something right-aligned once flex-wrap has dropped it onto a line of
// its own.
func PushEnd() Option {
	return func(r *rule) {
		r.pushEnd = true
	}
}

// CenterContent centers whatever the element contains, on both axes. A button
// holding nothing but an icon needs it: the icon is a replaced element with
// display: block, so the text-align a button carries by default does not move
// it and it sits against the leading edge.
func CenterContent() Option {
	return func(r *rule) {
		r.centerContent = true
	}
}

// Glyph colours what the element draws — its text and, through currentColor,
// its icons — with a surface's base colour, and leaves the background alone. It
// is the "tinted, not filled" treatment: a nav item that is merely available
// shows a coloured icon, the selected one gets the filled surface via As().
func Glyph(s Surface) Option {
	return func(r *rule) {
		r.hasGlyph = true
		r.glyph = s
	}
}

// ChipBox gives the element the shared chip width, the box a legend or a badge
// occupies. Fixing it is what makes a column of chips line up instead of each
// one hugging its own text; the text itself is truncated by the component that
// renders it.
func ChipBox() Option {
	return func(r *rule) {
		r.chipBox = true
	}
}

// Hide removes the element. Its use is inside On(): a part that exists on wide
// screens and not on a phone keeps its base styling and is switched off for the
// one device, which OnlyOn cannot express — OnlyOn hides by default and reveals
// per device, the opposite direction.
func Hide() Option {
	return func(r *rule) {
		r.hidden = true
	}
}

// ControlBox gives the element the shared control height, the rhythm every
// interactive row in the app is measured against — a list row, a form field.
// Pinning both to one token is what keeps them from drifting apart.
func ControlBox() Option {
	return func(r *rule) {
		r.controlBox = true
	}
}

// KeepSize does NOT reflow: maintains its size under any width.
func KeepSize() Option {
	return func(r *rule) {
		r.keepSize = true
	}
}

// EdgeToEdge has no border radius or margin: flush against parent container.
func EdgeToEdge() Option {
	return func(r *rule) {
		r.edgeToEdge = true
	}
}

// HideOverflow clips descendants (overflow: hidden).
func HideOverflow() Option {
	return func(r *rule) {
		r.hideOverflow = true
	}
}

// Visual options based on scales.

// Pad applies internal padding according to the space scale.
func Pad(s Space) Option {
	return func(r *rule) {
		r.hasPad = true
		r.pad = s
	}
}

// PadEdge pads one block edge only. Pad() is all four sides, which is the right
// default; this exists for the case where a fixed overlay covers the top of a
// panel and the content underneath has to start below it without gaining the
// same inset left and right.
func PadEdge(e Edge, s Space) Option {
	return func(r *rule) {
		r.hasPadEdge = true
		r.padEdge = e
		r.padEdgeSpace = s
	}
}

// Round applies border radius according to the radius scale.
func Round(rad Radius) Option {
	return func(r *rule) {
		r.hasRound = true
		r.round = rad
	}
}

// Raise applies shadow elevation according to the elevation scale.
func Raise(e Elevation) Option {
	return func(r *rule) {
		r.hasRaise = true
		r.raise = e
	}
}

// Width applies the relative width (Size) to the rule.
func Width(s Size) Option {
	return func(r *rule) {
		r.hasSize = true
		r.size = s
	}
}

// IconBox sizes a part as a square that never shrinks — the shape an icon needs.
// A bare <svg> with no width or height falls back to the replaced-element default
// of 300x150 and blows the layout apart, so every part that renders one must
// declare its box here.
func IconBox(s IconSize) Option {
	return func(r *rule) {
		r.hasIcon = true
		r.icon = s
	}
}

// As sets the surface decision (background, text, border, and radius default).
func As(s Surface) Option {
	return func(r *rule) {
		r.hasSurface = true
		r.surface = s
	}
}

// FontSize sets the text size.
func FontSize(ts TextSize) Option {
	return func(r *rule) {
		r.hasTextSize = true
		r.textSize = ts
	}
}

// FontWeight sets the font weight.
func FontWeight(w Weight) Option {
	return func(r *rule) {
		r.hasWeight = true
		r.weight = w
	}
}

// Animate applies a transition according to the motion scale.
func Animate(m Motion) Option {
	return func(r *rule) {
		r.hasMotion = true
		r.motion = m
	}
}
