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
