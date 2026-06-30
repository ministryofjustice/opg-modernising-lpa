package forms

// Field represents the basic information required to display a form field. It
// should be embedded in a more specific type.
type Field struct {
	// Name of the form field
	Name string

	// Label of the form field
	Label string

	// Input provided by the user, use Set instead of assigning to this
	Input string

	// Error, if any, from validating the input
	Error Error
}

// field returns itself to provide a way of getting the underlying data without
// needing to know the embedding type.
func (f Field) field() Field {
	return f
}
