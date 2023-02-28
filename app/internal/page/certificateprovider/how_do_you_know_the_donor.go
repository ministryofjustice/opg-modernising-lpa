package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howDoYouKnowTheDonorData struct {
	App    page.AppData
	Errors validation.List
	Donor  actor.Person
	Form   *howDoYouKnowTheDonorForm
}

func HowDoYouKnowTheDonor(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howDoYouKnowTheDonorData{
			App:   appData,
			Donor: lpa.You,
			Form: &howDoYouKnowTheDonorForm{
				How: lpa.CertificateProvider.DeclaredRelationship,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowTheDonorForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.DeclaredRelationship = data.Form.How

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.CertificateProvider.DeclaredRelationship == "personally" {
					return appData.Redirect(w, r, lpa, page.Paths.HowLongHaveYouKnownDonor)
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderYourDetails)
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowTheDonorForm struct {
	How string
}

func readHowDoYouKnowTheDonorForm(r *http.Request) *howDoYouKnowTheDonorForm {
	return &howDoYouKnowTheDonorForm{
		How: page.PostFormString(r, "how"),
	}
}

func (f *howDoYouKnowTheDonorForm) Validate() validation.List {
	var errors validation.List

	errors.String("how", "howYouKnowDonor", f.How,
		validation.Select("personally", "professionally"))

	return errors
}
