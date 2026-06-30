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

func ParsePostForm(r *http.Request, fields ...Parseable) []Field {
	var errors []Field
	r.ParseForm()

	for _, field := range fields {
		field.Parse(r.PostForm)
		if f := field.field(); f.Error != nil {
			errors = append(errors, f)
		}
	}

	return errors
}

type Form struct {
	Errors []Field
}

func (f *Form) ParsePostForm(r *http.Request, fields ...Parseable) bool {
	f.Errors = append(f.Errors, ParsePostForm(r, fields...)...)

	return len(f.Errors) == 0
}
