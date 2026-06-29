// Package forms provides validatable form fields and commonly used forms.
package forms

import (
	"net/http"
	"net/url"
)

type Parseable interface {
	field() Field
	Parse(url.Values)
}

type Form struct {
	Errors []Field
}

func (f *Form) ParsePostForm(r *http.Request, fields ...Parseable) bool {
	r.ParseForm()
	f.Errors = []Field{}

	for _, field := range fields {
		field.Parse(r.PostForm)
		if field := field.field(); field.Error != nil {
			f.Errors = append(f.Errors, field)
		}
	}

	return len(f.Errors) == 0
}
