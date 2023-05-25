package certificateprovider

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func LoginCallback(oneLoginClient OneLoginClient, sessionStore sesh.Store, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.CertificateProvider || oneLoginSession.Identity {
			return errors.New("certificate-provider callback with incorrect session")
		}

		idToken, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if err := sesh.SetCertificateProvider(sessionStore, r, w, &sesh.CertificateProviderSession{
			IDToken: idToken,
			Sub:     userInfo.Sub,
			Email:   userInfo.Email,
			LpaID:   oneLoginSession.LpaID,
		}); err != nil {
			return err
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
			SessionID: base64.StdEncoding.EncodeToString([]byte(userInfo.Sub)),
			LpaID:     oneLoginSession.LpaID,
		})

		_, err = certificateProviderStore.Create(ctx)
		if err != nil {
			var ccf *types.ConditionalCheckFailedException
			if !errors.As(err, &ccf) {
				return err
			}
		}

		return appData.Redirect(w, r, nil, page.Paths.CertificateProviderEnterDateOfBirth)
	}
}
