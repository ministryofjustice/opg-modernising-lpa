package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				How: lpa.CertificateProvider.Relationship,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.Relationship = data.Form.How

				requireLength := false
				if lpa.CertificateProvider.Relationship == "personally" {
					requireLength = true
				}

				if requireLength {
					// TODO: should stay as Completed if editing and not changing the answer here
					lpa.Tasks.CertificateProvider = actor.TaskInProgress
				} else {
					lpa.CertificateProvider.RelationshipLength = ""
					lpa.Tasks.CertificateProvider = actor.TaskCompleted
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if requireLength {
					return appData.Redirect(w, r, lpa, page.Paths.HowLongHaveYouKnownCertificateProvider.Format(lpa.ID))
				}

				return appData.Redirect(w, r, lpa, page.Paths.DoYouWantToNotifyPeople.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowYourCertificateProviderForm struct {
	How string
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	r.ParseForm()

	return &howDoYouKnowYourCertificateProviderForm{
		How: page.PostFormString(r, "how"),
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() validation.List {
	var errors validation.List

	errors.String("how", "howYouKnowCertificateProvider", f.How,
		validation.Select("personally", "professionally"))

	return errors
}
