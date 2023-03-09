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

type identityWithOneLoginCallbackData struct {
	App             page.AppData
	Errors          validation.List
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sesh.Store, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.CertificateProviderIdentityConfirmed() {
				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderReadTheLpa)
			} else {
				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderSelectYourIdentityOptions1)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if lpa.CertificateProviderIdentityConfirmed() {
			data.FullName = lpa.CertificateProviderIdentityUserData.FirstNames + " " + lpa.CertificateProviderIdentityUserData.LastName
			data.ConfirmedAt = lpa.CertificateProviderIdentityUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}
		if !oneLoginSession.CertificateProvider || !oneLoginSession.Identity {
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

		userData, err := oneLoginClient.ParseIdentityClaim(r.Context(), userInfo)
		if err != nil {
			return err
		}

		lpa.CertificateProviderIdentityUserData = userData

		if lpa.CertificateProviderIdentityConfirmed() {
			data.FullName = userData.FirstNames + " " + userData.LastName
			data.ConfirmedAt = userData.RetrievedAt

			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}
		} else {
			data.CouldNotConfirm = true
		}

		return tmpl(w, data)
	}
}
