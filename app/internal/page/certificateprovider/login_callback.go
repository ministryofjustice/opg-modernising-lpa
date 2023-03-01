package certificateprovider

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func LoginCallback(oneLoginClient OneLoginClient, sessionStore sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.CertificateProvider || oneLoginSession.Identity {
			return errors.New("certificate-provider callback with incorrect session")
		}

		accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if err := sesh.SetCertificateProvider(sessionStore, r, w, &sesh.CertificateProviderSession{
			Sub:            userInfo.Sub,
			Email:          userInfo.Email,
			LpaID:          oneLoginSession.LpaID,
			DonorSessionID: oneLoginSession.SessionID,
		}); err != nil {
			return err
		}

		return appData.Redirect(w, r, nil, page.Paths.CertificateProviderCheckYourName)
	}
}
