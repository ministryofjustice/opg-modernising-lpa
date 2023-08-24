package page

import (
	"net/http"
	"net/url"
	"strings"
)

func RouteToPrefix(prefix string, mux http.Handler, notFoundHandler Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 4)
		if len(parts) != 4 {
			notFoundHandler(AppDataFromContext(r.Context()), w, r)
			return
		}

		id, path := parts[2], "/"+parts[3]

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		if len(r.URL.RawPath) > len(prefix)+len(id) {
			r2.URL.RawPath = r.URL.RawPath[len(prefix)+len(id):]
		}

		mux.ServeHTTP(w, r2.WithContext(ContextWithSessionData(r2.Context(), &SessionData{LpaID: id})))
	}
}
