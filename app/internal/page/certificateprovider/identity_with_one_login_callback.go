package certificateprovider

import (
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithOneLoginCallbackData struct {
	App             page.AppData
	Errors          validation.List
	FirstNames      string
	LastName        string
	DateOfBirth     date.Date
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sesh.Store, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if certificateProvider.CertificateProviderIdentityConfirmed() {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProviderReadTheLpa)
			} else {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProviderSelectYourIdentityOptions1)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if certificateProvider.CertificateProviderIdentityConfirmed() {
			data.FirstNames = certificateProvider.IdentityUserData.FirstNames
			data.LastName = certificateProvider.IdentityUserData.LastName
			data.DateOfBirth = certificateProvider.IdentityUserData.DateOfBirth
			data.ConfirmedAt = certificateProvider.IdentityUserData.RetrievedAt

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

		_, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
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

		certificateProvider.IdentityUserData = userData

		if certificateProvider.CertificateProviderIdentityConfirmed() {
			data.FirstNames = userData.FirstNames
			data.LastName = userData.LastName
			data.DateOfBirth = userData.DateOfBirth
			data.ConfirmedAt = userData.RetrievedAt

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}
		} else {
			data.CouldNotConfirm = true
		}

		return tmpl(w, data)
	}
}
