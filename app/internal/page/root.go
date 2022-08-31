package page

import (
	"net/http"
)

func Root() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, startPath, http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	}
}
