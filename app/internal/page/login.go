package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
)

type LoginLogger interface {
	Print(v ...interface{})
}

type LoginOneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
}

func Login(logger LoginLogger, oneLoginClient LoginOneLoginClient, store sesh.Store, randomString func(int) string, redirect string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		locale := "en"
		if appData.Lang == localize.Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, false)

		if err := sesh.SetOneLogin(store, r, w, &sesh.OneLoginSession{
			State:    state,
			Nonce:    nonce,
			Locale:   locale,
			Redirect: redirect,
		}); err != nil {
			logger.Print(err)
			return nil
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
