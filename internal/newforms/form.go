package newforms

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
