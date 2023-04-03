package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                        page.AppData
	Errors                     validation.List
	CertificateProviderDetails actor.CertificateProvider
	Form                       *howDoYouKnowYourCertificateProviderForm
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howDoYouKnowYourCertificateProviderData{
			App:                        appData,
			CertificateProviderDetails: lpa.CertificateProviderDetails,
			Form: &howDoYouKnowYourCertificateProviderForm{
				Description: lpa.CertificateProviderDetails.RelationshipDescription,
				How:         lpa.CertificateProviderDetails.Relationship,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProviderDetails.Relationship = data.Form.How
				lpa.CertificateProviderDetails.RelationshipDescription = data.Form.Description

				requireLength := false

				if lpa.CertificateProviderDetails.Relationship != "legal-professional" && lpa.CertificateProviderDetails.Relationship != "health-professional" {
					requireLength = true
				}

				if requireLength {
					lpa.Tasks.CertificateProvider = page.TaskInProgress
				} else {
					lpa.CertificateProviderDetails.RelationshipLength = ""
					lpa.Tasks.CertificateProvider = page.TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if requireLength {
					return appData.Redirect(w, r, lpa, page.Paths.HowLongHaveYouKnownCertificateProvider)
				}

				return appData.Redirect(w, r, lpa, page.Paths.DoYouWantToNotifyPeople)
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
		Description: page.PostFormString(r, "description"),
		How:         page.PostFormString(r, "how"),
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
