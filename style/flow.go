//go:build !wasm

package style

// Stack define un ritmo vertical con hijos a ancho completo.
func Stack(gap Space) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "stack"
		r.FlowGap = gap
	}
}

// Row define un flujo horizontal que envuelve cuando no cabe.
func Row(gap Space) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "row"
		r.FlowGap = gap
	}
}

// Split define dos paneles que se apilan bajo su propio ancho.
func Split(ratio Ratio, gap Space) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "split"
		r.FlowRatio = ratio
		r.FlowGap = gap
	}
}

// Grid define auto-fit + minmax sin número fijo de columnas.
func Grid(min Track, gap Space) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "grid"
		r.FlowTrack = min
		r.FlowGap = gap
	}
}

// Center define una columna centrada con tope Prose/Content.
func Center() Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "center"
	}
}

// Cover llena el contenedor con un hijo centrado.
func Cover() Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "cover"
	}
}

// Reel define una tira horizontal con scroll-snap.
func Reel(gap Space) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "reel"
		r.FlowGap = gap
	}
}

// Frame define una caja de proporción fija (media).
func Frame(ratio Ratio) Opt {
	return func(r *Rule) {
		r.HasFlow = true
		r.FlowType = "frame"
		r.FlowRatio = ratio
	}
}
