package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginOneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
}

func Login(oneLoginClient LoginOneLoginClient, store sesh.Store, randomString func(int) string, redirect Path) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		locale := "en"
		if appData.Lang == localize.Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL, err := oneLoginClient.AuthCodeURL(state, nonce, locale, false)
		if err != nil {
			return err
		}

		if err := sesh.SetOneLogin(store, r, w, &sesh.OneLoginSession{
			State:    state,
			Nonce:    nonce,
			Locale:   locale,
			Redirect: redirect.Format(),
		}); err != nil {
			return err
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
