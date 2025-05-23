package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func IdentityWithOneLogin(oneLoginClient OneLoginClient, sessionStore SessionStore, randomString func(int) string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, _ *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		locale := ""
		if appData.Lang == localize.Cy {
			locale = "cy"
		}

		state := randomString(12)
		nonce := randomString(12)

		authCodeURL, err := oneLoginClient.AuthCodeURL(state, nonce, locale, onelogin.ConfidenceLevelLow)
		if err != nil {
			return err
		}

		if err := sessionStore.SetOneLogin(r, w, &sesh.OneLoginSession{
			State:    state,
			Nonce:    nonce,
			Locale:   locale,
			Redirect: certificateprovider.PathIdentityWithOneLoginCallback.Format(appData.LpaID),
		}); err != nil {
			return err
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
