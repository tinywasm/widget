package widget

// NameField is the shared anatomy of the form field.
//
// It is emitted by github.com/tinywasm/form (which builds the markup) and
// styled by github.com/tinywasm/components/fieldset (which is the global skin
// of the forms). It lives here because it crosses the boundary between these
// two libraries, and neither can own it without the other depending on it.
const NameField = Name("tw-field")

// Parts of the field. The names are generic by design: Part is local to its
// widget and only becomes a class through a Name, so "label" here produces
// "tw-field__label" and never collides with the "label" of another widget.
const (
	PartLabel      = Part("label")
	PartInput      = Part("input")
	PartError      = Part("error")
	PartRadioGroup = Part("radio-group")
)
