package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
				return page.Paths.CertificateProvider.ReadTheLpa.Redirect(w, r, appData, certificateProvider.LpaID)
			} else {
				return page.Paths.CertificateProvider.ProveYourIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
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

		oneLoginSession, err := sessionStore.OneLogin(r)
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

		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			data.FirstNames = userData.FirstNames
			data.LastName = userData.LastName
			data.DateOfBirth = userData.DateOfBirth
			data.ConfirmedAt = userData.RetrievedAt

			certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			if err := lpaStoreClient.SendCertificateProviderConfirmIdentity(r.Context(), lpa.LpaUID, certificateProvider); err != nil {
				return err
			}

		} else {
			data.CouldNotConfirm = true

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}
		}

		if userData.Status.IsFailed() {
			return page.Paths.CertificateProvider.UnableToConfirmIdentity.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		return tmpl(w, data)
	}
}
