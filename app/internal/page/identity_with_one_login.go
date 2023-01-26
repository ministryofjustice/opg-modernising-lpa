package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func IdentityWithOneLogin(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, randomString func(int) string) Handler {
	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   10 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		locale := ""
		if appData.Lang == Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, true)

		params := sessions.NewSession(store, "params")
		params.Values = map[interface{}]interface{}{
			"state":    state,
			"nonce":    nonce,
			"locale":   locale,
			"identity": true,
			"lpa-id":   appData.LpaID,
		}
		params.Options = cookieOptions

		if err := store.Save(r, w, params); err != nil {
			logger.Print(err)
			return nil
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
