package forms

import (
	"net/http"
	"net/url"
	"strings"
)

func makeRequest(query url.Values) *http.Request {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(query.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return r
}
