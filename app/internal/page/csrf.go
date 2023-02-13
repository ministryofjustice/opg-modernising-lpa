package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type contextKey string

func ValidateCsrf(next http.Handler, store sessions.Store, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		csrfSession, err := store.Get(r, "csrf")

		if r.Method == http.MethodPost {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if !csrfValid(r, csrfSession) {
				http.Error(w, "CSRF token not valid", http.StatusForbidden)
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
	formValue := r.PostFormValue("csrf")
	cookieValue, ok := csrfSession.Values["token"].(string)

	return ok && formValue == cookieValue
}
