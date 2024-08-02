// Package validation provides functionality to validate form data and present
// translatable errors.
package validation

type Localizer interface {
	Concat([]string, string) string
	Format(string, map[string]any) string
	FormatCount(string, int, map[string]interface{}) string
	T(string) string
}

type Field struct {
	Name  string
	Error FormattableError
}

type List []Field

func With(name string, error FormattableError) List {
	return List{Field{Name: name, Error: error}}
}

func (l List) None() bool {
	return len(l) == 0
}

func (l List) Any() bool {
	return len(l) > 0
}

func (l List) With(name string, error FormattableError) List {
	if l.Has(name) {
		return l
	}

	return append(l, Field{Name: name, Error: error})
}

func (l *List) Add(name string, error FormattableError) {
	if l.Has(name) {
		return
	}

	*l = append(*l, Field{Name: name, Error: error})
}

func (l *List) Append(other List) List {
	return append(*l, other...)
}

func (l List) Has(name string) bool {
	for _, field := range l {
		if field.Name == name {
			return true
		}
	}

	return false
}

func (l List) HasForDate(name, part string) bool {
	for _, field := range l {
		if field.Name == name {
			if err, ok := field.Error.(DateMissingError); ok {
				switch part {
				case "day":
					return err.MissingDay
				case "month":
					return err.MissingMonth
				case "year":
					return err.MissingYear
				}
			}

			return true
		}
	}

	return false
}

func (l List) Format(localizer Localizer, name string) string {
	for _, field := range l {
		if field.Name == name {
			return field.Error.Format(localizer)
		}
	}

	return ""
}
