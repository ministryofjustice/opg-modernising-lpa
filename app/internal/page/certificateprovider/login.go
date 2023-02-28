package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Login(logger Logger, oneLoginClient OneLoginClient, store sesh.Store, randomString func(int) string) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		locale := "en"
		if appData.Lang == localize.Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		sessionID := r.FormValue("sessionId")
		lpaID := r.FormValue("lpaId")

		authCodeURL := oneLoginClient.AuthCodeURL(state, nonce, locale, false)

		if err := sesh.SetOneLogin(store, r, w, &sesh.OneLoginSession{
			State:               state,
			Nonce:               nonce,
			Locale:              locale,
			CertificateProvider: true,
			Identity:            false,
			SessionID:           sessionID,
			LpaID:               lpaID,
		}); err != nil {
			logger.Print(err)
			return nil
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
