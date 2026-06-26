package forms

// validator defines the simple interface used for validating form values.
//
// We define validators and write the field types as:
//
//	type X { validators []validator[x] }
//
// instead of (the simpler):
//
//	type X { validators []func(x) Error }
//
// so that in the page tests we can use assert.Equals. Using bare funcs as the
// validators means they cannot be compared with reflect.DeepEquals, and so the
// tests will become much much more annoying. Even more annoying than having to
// define a type per validator...
type validator[T any] interface {
	Validate(v T) Error
}
