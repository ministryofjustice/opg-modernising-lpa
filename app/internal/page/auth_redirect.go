package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func AuthRedirect(logger Logger, oneLoginClient OneLoginClient, store sesh.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oneLoginSession, err := sesh.OneLogin(store, r)
		if err != nil {
			logger.Print(err)
			return
		}

		if oneLoginSession.State != r.FormValue("state") {
			logger.Print("state incorrect")
			return
		}

		lang := localize.En
		if oneLoginSession.Locale == "cy" {
			lang = localize.Cy
		}

		appData := AppData{Lang: lang, LpaID: oneLoginSession.LpaID}

		if oneLoginSession.CertificateProvider {
			appData.Redirect(w, r, nil, Paths.CertificateProviderLoginCallback+"?"+r.URL.RawQuery)
		} else if oneLoginSession.Identity {
			appData.Redirect(w, r, nil, Paths.IdentityWithOneLoginCallback+"?"+r.URL.RawQuery)
		} else {
			accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
			if err != nil {
				logger.Print(err)
				return
			}

			userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
			if err != nil {
				logger.Print(err)
				return
			}

			if err := sesh.SetDonor(store, r, w, &sesh.DonorSession{
				Sub:   userInfo.Sub,
				Email: userInfo.Email,
			}); err != nil {
				logger.Print(err)
				return
			}

			appData.Redirect(w, r, nil, Paths.Dashboard)
		}
	}
}
