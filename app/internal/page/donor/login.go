package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Login(logger page.Logger, oneLoginClient page.OneLoginClient, store sesh.Store, randomString func(int) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locale := "en"

		if r.URL.Query().Has("locale") {
			locale = r.URL.Query().Get("locale")
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, false)

		if err := sesh.SetOneLoginSession(store, r, w, &sesh.OneLoginSession{
			State:  state,
			Nonce:  nonce,
			Locale: locale,
		}); err != nil {
			logger.Print(err)
			return
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
	}
}
