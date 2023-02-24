package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func LoginCallback(oneLoginClient OneLoginClient, store sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(store, r)
		if err != nil {
			return err
		}

		accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if err := sesh.SetDonor(store, r, w, &sesh.DonorSession{
			Sub:   userInfo.Sub,
			Email: userInfo.Email,
		}); err != nil {
			return err
		}

		return appData.Redirect(w, r, nil, page.Paths.Dashboard)
	}
}
