package forms

// Field represents the basic information required to display a form field. It
// should be embedded in a more specific type.
type Field struct {
	Name  string // Name of the form field
	Label string // Label of the form field
	Input string // Input provided by the user, trimmed of spaces
	Error Error  // Error, if any, from validating the input
}

// field returns itself to provide a way of getting the underlying data without
// needing to know the embedding type.
func (f Field) field() Field {
	return f
}
