package forms

type Localizer interface {
	T(msgid string) string
	Format(msgid string, data map[string]any) string
}

type withError[T any] struct {
	replace Error
	wrapped validator[T]
}

func (v withError[T]) Validate(t T) Error {
	if v.wrapped.Validate(t) != nil {
		return v.replace
	}
	return nil
}

type withErrorLabel[T any] struct {
	replace string
	wrapped validator[T]
}

func (v withErrorLabel[T]) Validate(t T) Error {
	if error := v.wrapped.Validate(t); error != nil {
		if ferror, ok := error.(formattedError); ok {
			ferror.Data["Label"] = v.replace
			return ferror
		}

		return error
	}
	return nil
}

type Error interface {
	Format(localizer Localizer) string
}

type ErrorMessage string

func (e ErrorMessage) Format(l Localizer) string {
	return l.T(string(e))
}

type formattedError struct {
	Key  string
	Data map[string]any
}

func (e formattedError) Format(l Localizer) string {
	return l.Format(e.Key, e.Data)
}

func newEmptyError(label string) formattedError {
	return formattedError{
		Key:  "errorEnter",
		Data: map[string]any{"Label": label},
	}
}

func newTooLongError(label string, length int) formattedError {
	return formattedError{
		Key: "errorStringTooLong",
		Data: map[string]any{
			"Label":  label,
			"Length": length,
		},
	}
}

func newSelectError(label string) formattedError {
	return formattedError{
		Key:  "errorSelect",
		Data: map[string]any{"Label": label},
	}
}

func newPhoneError(label string) formattedError {
	return formattedError{
		Key:  "errorPhone",
		Data: map[string]any{"Label": label},
	}
}
