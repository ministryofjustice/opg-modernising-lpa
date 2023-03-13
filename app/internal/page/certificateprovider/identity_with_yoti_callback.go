package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
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

func IdentityWithYotiCallback(tmpl template.Template, yotiClient YotiClient, lpaStore LpaStore) page.Handler {
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

		data := &identityWithYotiCallbackData{App: appData}

		if lpa.CertificateProviderIdentityConfirmed() {
			data.FirstNames = lpa.CertificateProviderIdentityUserData.FirstNames
			data.LastName = lpa.CertificateProviderIdentityUserData.LastName
			data.DateOfBirth = lpa.CertificateProviderIdentityUserData.DateOfBirth
			data.ConfirmedAt = lpa.CertificateProviderIdentityUserData.RetrievedAt

			return tmpl(w, data)
		}

		user, err := yotiClient.User(r.FormValue("token"))
		if err != nil {
			return err
		}

		lpa.CertificateProviderIdentityUserData = user

		if lpa.CertificateProviderIdentityConfirmed() {
			if err := lpaStore.Put(r.Context(), lpa); err != nil {
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
