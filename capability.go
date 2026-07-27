package widget

// Widget es la identidad. Es lo único obligatorio.
type Widget interface {
	WidgetName() Name
	WidgetKind() Kind
}

// Capacidades — cada costura asevera solo la que necesita (patrón de la casa,
// el mismo de view.Saver / view.Deleter).
type Selectable interface{ Select(id string) }
type Dismissible interface{ Dismiss() }
type Expandable interface{ Expand(open bool) }
