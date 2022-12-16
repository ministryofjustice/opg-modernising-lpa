package page

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
)

type loginClient interface {
	AuthCodeURL(state, nonce, locale string) string
}

type State struct {
	Uid    string `json:"uid"`
	Locale string `json:"locale"`
}

func Login(logger Logger, c loginClient, store sessions.Store, secure bool, randomString func(int) string) http.HandlerFunc {
	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   10 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   secure,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		locale := "en"

		if r.URL.Query().Has("locale") {
			locale = r.URL.Query().Get("locale")
		}

		state := State{
			Uid:    randomString(12),
			Locale: locale,
		}

		j, _ := json.Marshal(state)
		encodedState := base64.StdEncoding.EncodeToString(j)

		nonce := randomString(12)

		authCodeURL := c.AuthCodeURL(encodedState, nonce, locale)

		params := sessions.NewSession(store, "params")
		params.Values = map[interface{}]interface{}{
			"state": encodedState,
			"nonce": nonce,
		}
		params.Options = cookieOptions

		if err := store.Save(r, w, params); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
