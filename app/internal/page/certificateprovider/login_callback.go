package certificateprovider

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
)

func LoginCallback(oneLoginClient OneLoginClient, sessionStore sesh.Store, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}

		idToken, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{
			IDToken: idToken,
			Sub:     userInfo.Sub,
			Email:   userInfo.Email,
		}); err != nil {
			return err
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
			SessionID: base64.StdEncoding.EncodeToString([]byte(userInfo.Sub)),
			LpaID:     oneLoginSession.LpaID,
		})

		_, err = certificateProviderStore.Create(ctx, oneLoginSession.SessionID)
		if err != nil {
			var ccf *types.ConditionalCheckFailedException
			if !errors.As(err, &ccf) {
				return err
			}
		}

		appData.LpaID = oneLoginSession.LpaID
		return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.EnterDateOfBirth)
	}
}
