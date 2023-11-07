package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
	Options             actor.CertificateProviderRelationshipOptions
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				How: lpa.CertificateProvider.Relationship,
			},
			Options: actor.CertificateProviderRelationshipValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.Relationship = data.Form.How

				requireLength := false
				if lpa.CertificateProvider.Relationship.IsPersonally() {
					requireLength = true
				}

				if !requireLength {
					lpa.CertificateProvider.RelationshipLength = 0
				}

				lpa.Tasks.CertificateProvider = actor.TaskInProgress

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if requireLength {
					return appData.Redirect(w, r, lpa, page.Paths.HowLongHaveYouKnownCertificateProvider.Format(lpa.ID))
				}

				return appData.Redirect(w, r, lpa, page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowYourCertificateProviderForm struct {
	How   actor.CertificateProviderRelationship
	Error error
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	how, err := actor.ParseCertificateProviderRelationship(page.PostFormString(r, "how"))

	return &howDoYouKnowYourCertificateProviderForm{
		How:   how,
		Error: err,
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.Error("how", "howYouKnowCertificateProvider", f.Error,
		validation.Selected())

	return errors
}
