package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func IdentityWithOneLogin(oneLoginClient OneLoginClient, sessionStore SessionStore, randomString func(int) string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		locale := ""
		if appData.Lang == localize.Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL, err := oneLoginClient.AuthCodeURL(state, nonce, locale, true)
		if err != nil {
			return err
		}

		if err := sessionStore.SetOneLogin(r, w, &sesh.OneLoginSession{
			State:    state,
			Nonce:    nonce,
			Locale:   locale,
			LpaID:    donor.LpaID,
			Redirect: page.Paths.IdentityWithOneLoginCallback.Format(donor.LpaID),
		}); err != nil {
			return err
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
