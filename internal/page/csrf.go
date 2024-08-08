package page

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type contextKey string

var ErrCsrfInvalid = errors.New("CSRF token not valid")

const csrfTokenLength = 12

func ValidateCsrf(next http.Handler, store SessionStore, randomString func(int) string, errorHandler ErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		csrfSession, err := store.Csrf(r)

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

		if err != nil {
			csrfSession = &sesh.CsrfSession{Token: randomString(csrfTokenLength)}
			_ = store.SetCsrf(r, w, csrfSession)
		}

		appData := appcontext.DataFromContext(ctx)
		appData.CsrfToken = csrfSession.Token

		next.ServeHTTP(w, r.WithContext(appcontext.ContextWithData(ctx, appData)))
	}
}

func csrfValid(r *http.Request, csrfSession *sesh.CsrfSession) bool {
	cookieValue := csrfSession.Token

	if mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err == nil && mediaType == "multipart/form-data" {
		var buf bytes.Buffer
		reader := multipart.NewReader(io.TeeReader(r.Body, &buf), params["boundary"])

		part, err := reader.NextPart()
		if err != nil {
			return false
		}

		if part.FormName() != "csrf" {
			return false
		}

		lmt := io.LimitReader(part, csrfTokenLength+1)
		value, _ := io.ReadAll(lmt)

		r.Body = newMultiReadCloser(io.NopCloser(&buf), r.Body)
		return string(value) == cookieValue
	}

	return r.PostFormValue("csrf") == cookieValue
}
