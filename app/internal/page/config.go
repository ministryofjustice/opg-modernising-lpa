package page

import "net/http"

type RumConfig struct {
	GuestRoleArn      string
	Endpoint          string
	ApplicationRegion string
	IdentityPoolID    string
	ApplicationID     string
}

func CacheControlHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=2592000")
		h.ServeHTTP(w, r)
	})
}
