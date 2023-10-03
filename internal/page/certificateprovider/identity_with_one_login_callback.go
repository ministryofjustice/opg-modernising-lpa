package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sesh.Store, certificateProviderStore CertificateProviderStore, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.ReadTheLpa.Format(certificateProvider.LpaID))
			} else {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.SelectYourIdentityOptions1.Format(certificateProvider.LpaID))
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
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
		certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
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
