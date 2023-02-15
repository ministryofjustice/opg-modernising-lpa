package certificateprovider

import (
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type loginCallbackData struct {
	App             page.AppData
	Errors          validation.List
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func LoginCallback(tmpl template.Template, oneLoginClient page.OneLoginClient, sessionStore sesh.Store, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			certificateProviderSession, err := sesh.CertificateProvider(sessionStore, r)
			if err != nil {
				return err
			}

			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
				SessionID: certificateProviderSession.DonorSessionID,
				LpaID:     certificateProviderSession.LpaID,
			})

			lpa, err := lpaStore.Get(ctx)
			if err != nil {
				return err
			}

			if lpa.CertificateProviderUserData.OK {
				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderYourDetails)
			} else {
				return appData.Redirect(w, r, lpa, page.Paths.Start)
			}
		}

		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.CertificateProvider || !oneLoginSession.Identity {
			return errors.New("certificate-provider callback with incorrect session")
		}

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
			SessionID: oneLoginSession.SessionID,
			LpaID:     oneLoginSession.LpaID,
		})

		lpa, err := lpaStore.Get(ctx)
		if err != nil {
			return err
		}

		data := &loginCallbackData{App: appData}

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

		if lpa.CertificateProviderUserData.OK {
			data.FullName = lpa.CertificateProviderUserData.FullName
			data.ConfirmedAt = lpa.CertificateProviderUserData.RetrievedAt

			if err := sesh.SetCertificateProvider(sessionStore, r, w, &sesh.CertificateProviderSession{
				Sub:            userInfo.Sub,
				Email:          userInfo.Email,
				LpaID:          oneLoginSession.LpaID,
				DonorSessionID: oneLoginSession.SessionID,
			}); err != nil {
				return err
			}

			return tmpl(w, data)
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

			if err := sesh.SetCertificateProvider(sessionStore, r, w, &sesh.CertificateProviderSession{
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
