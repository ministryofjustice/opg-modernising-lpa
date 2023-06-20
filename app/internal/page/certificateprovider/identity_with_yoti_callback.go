package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type identityWithYotiCallbackData struct {
	App             page.AppData
	Errors          validation.List
	FirstNames      string
	LastName        string
	DateOfBirth     date.Date
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithYotiCallback(tmpl template.Template, yotiClient YotiClient, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if certificateProvider.CertificateProviderIdentityConfirmed() {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.ReadTheLpa.Format(certificateProvider.LpaID))
			} else {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.SelectYourIdentityOptions1.Format(certificateProvider.LpaID))
			}
		}

		data := &identityWithYotiCallbackData{App: appData}

		if certificateProvider.CertificateProviderIdentityConfirmed() {
			data.FirstNames = certificateProvider.IdentityUserData.FirstNames
			data.LastName = certificateProvider.IdentityUserData.LastName
			data.DateOfBirth = certificateProvider.IdentityUserData.DateOfBirth
			data.ConfirmedAt = certificateProvider.IdentityUserData.RetrievedAt

			return tmpl(w, data)
		}

		user, err := yotiClient.User(r.FormValue("token"))
		if err != nil {
			return err
		}

		certificateProvider.IdentityUserData = user

		if certificateProvider.CertificateProviderIdentityConfirmed() {
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			data.FirstNames = user.FirstNames
			data.LastName = user.LastName
			data.DateOfBirth = user.DateOfBirth
			data.ConfirmedAt = user.RetrievedAt
		} else {
			data.CouldNotConfirm = true
		}

		return tmpl(w, data)
	}
}
