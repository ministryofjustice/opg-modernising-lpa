package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

func IdentityWithOneLogin(oneLoginClient OneLoginClient, sessionStore SessionStore, randomString func(int) string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, _ *voucherdata.Provided) error {
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
			Redirect: voucher.PathIdentityWithOneLoginCallback.Format(appData.LpaID),
		}); err != nil {
			return err
		}

		http.Redirect(w, r, authCodeURL, http.StatusFound)
		return nil
	}
}
