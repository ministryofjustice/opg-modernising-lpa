package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type provideCertificateData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider *actor.CertificateProviderProvidedDetails
	Form                *provideCertificateForm
}

func ProvideCertificate(tmpl template.Template, donorStore DonorStore, now func() time.Time, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if lpa.Submitted.IsZero() {
			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderStart)
		}

		data := &provideCertificateData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Form: &provideCertificateForm{
				AgreeToStatement: certificateProvider.Certificate.AgreeToStatement,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readProvideCertificateForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				certificateProvider.Certificate.AgreeToStatement = true
				certificateProvider.Certificate.Agreed = now()
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProvided)
			}
		}

		return tmpl(w, data)
	}
}

type provideCertificateForm struct {
	AgreeToStatement bool
}

func readProvideCertificateForm(r *http.Request) *provideCertificateForm {
	return &provideCertificateForm{
		AgreeToStatement: page.PostFormString(r, "agree-to-statement") == "1",
	}
}

func (f *provideCertificateForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("agree-to-statement", "toSignAsCertificateProvider", f.AgreeToStatement,
		validation.Selected())

	return errors
}
