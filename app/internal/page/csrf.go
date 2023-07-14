package page

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type contextKey string

var ErrCsrfInvalid = errors.New("CSRF token not valid")

func ValidateCsrf(next http.Handler, store sessions.Store, randomString func(int) string, errorHandler ErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		csrfSession, err := store.Get(r, "csrf")

		if r.Method == http.MethodPost {
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !csrfValid(r, csrfSession) {
				errorHandler(w, r, ErrCsrfInvalid)
				return
			}
		}

		if csrfSession.IsNew {
			csrfSession.Values = map[any]any{"token": randomString(12)}
			csrfSession.Options = &sessions.Options{
				MaxAge:   24 * 60 * 60,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			_ = store.Save(r, w, csrfSession)
		}

		appData := AppDataFromContext(ctx)
		appData.CsrfToken, _ = csrfSession.Values["token"].(string)

		next.ServeHTTP(w, r.WithContext(ContextWithAppData(ctx, appData)))
	}
}

func csrfValid(r *http.Request, csrfSession *sessions.Session) bool {
	cookieValue, ok := csrfSession.Values["token"].(string)
	if !ok {
		return false
	}

	if contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";"); contentType == "multipart/form-data" {
		// for multipart/form-data requests the csrf token must be checked where the
		// other fields are read as we can't read the body twice
		return true
	}

	return r.PostFormValue("csrf") == cookieValue
}
