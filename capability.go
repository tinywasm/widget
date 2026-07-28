package widget

// Widget is the identity. It is the only mandatory interface.
type Widget interface {
	WidgetName() Name
	WidgetKind() Kind
}

// Capabilities — each hook asserts only the capability it needs.
type Selectable interface{ Select(id string) }
type Dismissible interface{ Dismiss() }
type Expandable interface{ Expand(open bool) }
