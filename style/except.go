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
