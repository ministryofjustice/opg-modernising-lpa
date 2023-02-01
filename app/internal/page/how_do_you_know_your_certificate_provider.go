package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 AppData
	Errors              validation.List
	CertificateProvider CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				Description: lpa.CertificateProvider.RelationshipDescription,
				How:         lpa.CertificateProvider.Relationship,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.Relationship = data.Form.How
				lpa.CertificateProvider.RelationshipDescription = data.Form.Description

				requireLength := false

				if lpa.CertificateProvider.Relationship != "legal-professional" && lpa.CertificateProvider.Relationship != "health-professional" {
					requireLength = true
				}

				if requireLength {
					lpa.Tasks.CertificateProvider = TaskInProgress
				} else {
					lpa.CertificateProvider.RelationshipLength = ""
					lpa.Tasks.CertificateProvider = TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if requireLength {
					return appData.Redirect(w, r, lpa, Paths.HowLongHaveYouKnownCertificateProvider)
				}

				return appData.Redirect(w, r, lpa, Paths.CheckYourLpa)
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowYourCertificateProviderForm struct {
	Description string
	How         string
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	r.ParseForm()

	return &howDoYouKnowYourCertificateProviderForm{
		Description: postFormString(r, "description"),
		How:         postFormString(r, "how"),
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.String("how", "howYouKnowCertificateProvider", f.How,
		validation.Select("friend", "neighbour", "colleague", "health-professional", "legal-professional", "other"))

	if f.How == "other" {
		errors.String("description", "description", f.Description,
			validation.Empty())
	}

	return errors
}
