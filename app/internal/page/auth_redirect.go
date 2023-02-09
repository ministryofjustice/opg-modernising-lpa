package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func AuthRedirect(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, secure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		oneLoginSession, err := getOneLoginSession(store, r)
		if err != nil {
			logger.Print(err)
			return
		}

		if oneLoginSession.State != r.FormValue("state") {
			logger.Print("state incorrect")
			return
		}

		lang := En
		if oneLoginSession.Locale == "cy" {
			lang = Cy
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

			if err := setDonorSession(store, r, w, &DonorSession{
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
