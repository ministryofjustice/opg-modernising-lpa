package forms

type Localizer interface {
	Format(msgid string, data map[string]any) string
}

type Error interface {
	Format(localizer Localizer) string
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
