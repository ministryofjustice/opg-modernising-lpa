package validation

type Field struct {
	Name  string
	Error string
}

type List []Field

func With(name, error string) List {
	return List{Field{Name: name, Error: error}}
}

func (l List) None() bool {
	return len(l) == 0
}

func (l List) Any() bool {
	return len(l) > 0
}

func (l List) With(name, error string) List {
	if l.Has(name) {
		return l
	}

	return append(l, Field{Name: name, Error: error})
}

func (l *List) Add(name, error string) {
	if l.Has(name) {
		return
	}

	*l = append(*l, Field{Name: name, Error: error})
}

func (l List) Get(name string) string {
	for _, field := range l {
		if field.Name == name {
			return field.Error
		}
	}

	return ""
}

func (l List) Has(name string) bool {
	for _, field := range l {
		if field.Name == name {
			return true
		}
	}

	return false
}
