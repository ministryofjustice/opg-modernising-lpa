package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterTrustCorporationData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterTrustCorporationForm
	LpaID  string
}

func EnterTrustCorporation(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		trustCorporation := donor.Attorneys.TrustCorporation

		data := &enterTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:          trustCorporation.Name,
				CompanyNumber: trustCorporation.CompanyNumber,
				Email:         trustCorporation.Email,
			},
			LpaID: donor.LpaID,
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
				trustCorporation.CompanyNumber = data.Form.CompanyNumber
				trustCorporation.Email = data.Form.Email
				donor.Attorneys.TrustCorporation = trustCorporation

				donor.Tasks.ChooseAttorneys = page.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.EnterTrustCorporationAddress.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type enterTrustCorporationForm struct {
	Name          string
	CompanyNumber string
	Email         string
}

func readEnterTrustCorporationForm(r *http.Request) *enterTrustCorporationForm {
	return &enterTrustCorporationForm{
		Name:          page.PostFormString(r, "name"),
		CompanyNumber: page.PostFormString(r, "company-number"),
		Email:         page.PostFormString(r, "email"),
	}
}

func (f *enterTrustCorporationForm) Validate() validation.List {
	var errors validation.List

	errors.String("name", "companyName", f.Name,
		validation.Empty())

	errors.String("company-number", "companyNumber", f.CompanyNumber,
		validation.Empty())

	errors.String("email", "companyEmailAddress", f.Email,
		validation.Email())

	return errors
}
