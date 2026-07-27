//go:build !wasm

package style

// Fill toma el alto disponible completo.
func Fill() Opt {
	return func(r *Rule) {
		r.Fill = true
	}
}

// Scrolls desborda internamente en vez de crecer. Implica Fill().
func Scrolls() Opt {
	return func(r *Rule) {
		r.Scrolls = true
		r.Fill = true
	}
}

// Fixed NO reflota: mantiene su disposición bajo cualquier ancho.
func Fixed() Opt {
	return func(r *Rule) {
		r.Fixed = true
	}
}

// Flush es sin radio ni margen: pega a ras del contenedor padre.
func Flush() Opt {
	return func(r *Rule) {
		r.Flush = true
	}
}

// Clip recorta a los hijos (overflow:hidden).
func Clip() Opt {
	return func(r *Rule) {
		r.Clip = true
	}
}

// Opciones visuales basadas en escalas.

// Pad aplica relleno interno (padding) según la escala de espacio.
func Pad(s Space) Opt {
	return func(r *Rule) {
		r.HasPad = true
		r.Pad = s
	}
}

// Round aplica radio de borde (border-radius) según la escala de radio.
func Round(rad Radius) Opt {
	return func(r *Rule) {
		r.HasRound = true
		r.Round = rad
	}
}

// Raise aplica elevación (box-shadow) según la escala de elevación.
func Raise(e Elevation) Opt {
	return func(r *Rule) {
		r.HasRaise = true
		r.Raise = e
	}
}

// Width aplica la única medida de ancho relativa (Size) aceptada en la API.
func Width(s Size) Opt {
	return func(r *Rule) {
		r.HasSize = true
		r.Size = s
	}
}

// On establece la decisión de superficie de color (background, texto y borde).
func On(s Surface) Opt {
	return func(r *Rule) {
		r.HasSurface = true
		r.Surface = s
	}
}

// Text establece el tamaño del texto.
func Text(ts TextSize) Opt {
	return func(r *Rule) {
		r.HasTextSize = true
		r.TextSize = ts
	}
}

// FontWeight establece el grosor de la fuente.
func FontWeight(w Weight) Opt {
	return func(r *Rule) {
		r.HasWeight = true
		r.Weight = w
	}
}
