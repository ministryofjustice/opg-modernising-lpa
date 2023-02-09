package page

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderLoginCallbackData struct {
	App             AppData
	Errors          validation.List
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func CertificateProviderLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sessions.Store, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			certificateProviderSession, err := getCertificateProviderSession(sessionStore, r)
			if err != nil {
				return err
			}

			ctx := contextWithSessionData(r.Context(), &sessionData{
				SessionID: certificateProviderSession.DonorSessionID,
				LpaID:     certificateProviderSession.LpaID,
			})

			lpa, err := lpaStore.Get(ctx)
			if err != nil {
				return err
			}

			if lpa.CertificateProviderUserData.OK {
				return appData.Redirect(w, r, lpa, Paths.CertificateProviderYourDetails)
			} else {
				return appData.Redirect(w, r, lpa, Paths.Start)
			}
		}

		oneLoginSession, err := getOneLoginSession(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.CertificateProvider || !oneLoginSession.Identity {
			return errors.New("certificate-provider callback with incorrect session")
		}

		ctx := contextWithSessionData(r.Context(), &sessionData{
			SessionID: oneLoginSession.SessionID,
			LpaID:     oneLoginSession.LpaID,
		})

		lpa, err := lpaStore.Get(ctx)
		if err != nil {
			return err
		}

		data := &certificateProviderLoginCallbackData{App: appData}

		if lpa.CertificateProviderUserData.OK {
			data.FullName = lpa.CertificateProviderUserData.FullName
			data.ConfirmedAt = lpa.CertificateProviderUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		accessToken, err := oneLoginClient.Exchange(ctx, r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(ctx, accessToken)
		if err != nil {
			return err
		}

		userData, err := oneLoginClient.ParseIdentityClaim(ctx, userInfo)
		if err != nil {
			return err
		}

		if !userData.OK {
			data.CouldNotConfirm = true
		} else {
			lpa.CertificateProviderUserData = userData

			if err := lpaStore.Put(ctx, lpa); err != nil {
				return err
			}

			if err := setCertificateProviderSession(sessionStore, r, w, &CertificateProviderSession{
				Sub:            userInfo.Sub,
				Email:          userInfo.Email,
				LpaID:          oneLoginSession.LpaID,
				DonorSessionID: oneLoginSession.SessionID,
			}); err != nil {
				return err
			}

			data.FullName = userData.FullName
			data.ConfirmedAt = userData.RetrievedAt
		}

		return tmpl(w, data)
	}
}
