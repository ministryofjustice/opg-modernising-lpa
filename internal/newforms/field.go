package newforms

type Field struct {
	Name  string // Name of the form field
	Label string // Label of the form field, used for constructing errors
	Input string // Input provided by the user, trimmed of spaces
	Error Error  // Error, if any, from validating the input
}

func (f Field) field() Field {
	return f
}
