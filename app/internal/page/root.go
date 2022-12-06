package page

import (
	"net/http"
)

func Root(paths AppPaths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, paths.Start, http.StatusFound)
		} else {
			http.NotFound(w, r)
		}
	}
}
