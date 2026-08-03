//go:build !wasm

package style

type flowType uint8

const (
	flowNone flowType = iota
	flowStack
	flowRow
	flowSplit
	flowGrid
	flowCenter
	flowFillCentered
	flowScrollRow
	flowMediaBox
	flowCover
	flowSidebar
	flowMasterDetail
	flowSlideDeck
)

// Stack defines a vertical rhythm with children at full width.
func Stack(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowStack
		r.flowGap = gap
	}
}

// Row defines a horizontal flow that wraps when it does not fit.
func Row(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowRow
		r.flowGap = gap
	}
}

// Split defines two panels that stack below their own width.
func Split(ratio SplitRatio, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSplit
		r.flowRatio = ratio
		r.flowGap = gap
	}
}

// Grid defines auto-fit + minmax without a fixed number of columns.
func Grid(min ColumnWidth, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowGrid
		r.flowWidth = min
		r.flowGap = gap
	}
}

// Center defines a centered column with an optional maximum size (defaults to Readable).
func Center(max ...Size) Option {
	sz := Readable
	if len(max) > 0 {
		sz = max[0]
	}
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowCenter
		r.hasSize = true
		r.size = sz
	}
}

// FillCentered fills the container with a centered child.
func FillCentered() Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowFillCentered
	}
}

// ScrollRow defines a horizontal scrolling strip with scroll-snap.
func ScrollRow(gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowScrollRow
		r.flowGap = gap
	}
}

// MediaBox defines a box of fixed aspect ratio.
func MediaBox(a Aspect) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowMediaBox
		r.flowAspect = a
	}
}

// Cover locks the frame to the viewport height and stacks its children
// vertically. It is the outermost frame of an application shell: use KeepSize()
// on the children that must not shrink (a header) and Fill() on the one that
// takes the remaining height. Do not nest one Cover inside another.
//
// The height is definite, not a floor, so a Fill() descendant resolves against
// it and a HideOverflow() or Scroll() descendant actually clips. A shell whose
// content can exceed the viewport must therefore give that descendant Scroll();
// otherwise the overflow is unreachable.
func Cover() Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowCover
	}
}

// SlideDeck apila a sus hijos en capas que ocupan el contenedor entero y muestra
// solo aquel que lleva el estado widget.Current; los demás quedan aparcados en el
// borde inline-start y entran deslizándose de izquierda a derecha cuando les toca.
//
// Es la forma de cambiar de panel en un shell SIN crear un scroller: un contenedor
// de scroll-snap horizontal aquí encadena con el scroll-snap horizontal que un
// módulo pueda tener adentro, y el gesto de deslizar dentro del contenido termina
// cambiando de sección sola.
//
// Todos los hijos siguen montados en el DOM. Ese es el trato: el estado decide
// cuál está en pantalla, nadie desmonta nada. El contenedor es el bloque contenedor
// de sus hijos, así que un Docked(Parent) dentro de un panel se resuelve contra SU
// panel — no hace falta Anchor() en el hijo, y ponerlo lo ROMPE: el position:
// relative de @layer widgets gana sobre el position: absolute que este flujo emite
// en @layer primitives.
//
// m gobierna la duración del deslizamiento. MotionNone conmuta sin animación.
func SlideDeck(m Motion) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSlideDeck
		r.flowMotion = m
	}
}

// MasterDetail turns a two-panel container into a horizontal scroll-snap strip
// for a narrow screen: the master list rests where the browser's default scroll
// position already is, and the detail sits beside it at `detail` of the strip's
// width, so a sliver of the list stays visible and the panel it came from is
// obvious. Swiping is a native scroll; snapping to the detail is a plain
// ScrollIntoView from the row handler.
//
// The FIRST TWO element children are the panels, in the same DOM order a desktop
// Split uses: detail first, master second. Anything after them — a modal mount
// point, a portal anchor — is left alone, which is why this addresses them by
// position and not with :first-child/:last-child. The strip is laid out RTL so
// the master — the second child, given order 1 — lands at the start edge, which
// RTL puts on the right, exactly where scroll position 0 already rests. That is
// what removes the need for a scroll nudge at mount time, which this framework's
// component contract has no hook for. Each panel resets to LTR so only the
// outer strip's flow is mirrored, never the content.
func MasterDetail(detail Size) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowMasterDetail
		r.flowDetail = detail
	}
}

// Sidebar places a fixed-width rail beside a fluid content area. The rail keeps
// its width; the content takes everything else. Below the point where the content
// can no longer hold its minimum width the two reflow into a single column, with
// no media query involved.
//
// The container MUST have exactly two element children. Which one is the rail is
// decided by side, not by DOM order: SideEnd makes the LAST child the rail.
func Sidebar(side Side, width RailWidth, gap Space) Option {
	return func(r *rule) {
		r.hasFlow = true
		r.flowType = flowSidebar
		r.flowSide = side
		r.flowRail = width
		r.flowGap = gap
	}
}
