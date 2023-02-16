package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type provideCertificateData struct {
	App         page.AppData
	Errors      validation.List
	Certificate page.Certificate
	Form        *provideCertificateForm
}

func ProvideCertificate(tmpl template.Template, lpaStore page.LpaStore, now func() time.Time) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if lpa.Submitted.IsZero() {
			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderStart)
		}

		data := &provideCertificateData{
			App:         appData,
			Certificate: lpa.Certificate,
			Form: &provideCertificateForm{
				AgreeToStatement: lpa.Certificate.AgreeToStatement,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readProvideCertificateForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.Certificate.AgreeToStatement = true
				lpa.Certificate.Agreed = now()
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
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

	errors.Bool("agree-to-statement", "agreeToStatement", f.AgreeToStatement,
		validation.Selected())

	return errors
}
