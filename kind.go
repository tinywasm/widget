package widget

// Kind es el tipo de widget según WAI-ARIA Authoring Practices. Determina el rol,
// los estados válidos y el teclado esperado. Cerrado a propósito: si un widget no
// encaja en ninguno, casi siempre es que son dos widgets.
type Kind uint8

const (
	Region Kind = iota // contenedor sin semántica de interacción
	Listbox            // targetlist
	Menu               // el menú ⋮
	Dialog             // modaldialog
	Disclosure         // <details> desplegable
	Tabs
	Toolbar
	Grid
	Combobox
	Form
	Alert              // toasts de platformd
)
