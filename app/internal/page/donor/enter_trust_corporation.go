package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type enterTrustCorporationData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterTrustCorporationForm
	LpaID  string
}

func EnterTrustCorporation(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &enterTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:          lpa.TrustCorporation.Name,
				CompanyNumber: lpa.TrustCorporation.CompanyNumber,
				Email:         lpa.TrustCorporation.Email,
			},
			LpaID: lpa.ID,
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.TrustCorporation.Name = data.Form.Name
				lpa.TrustCorporation.CompanyNumber = data.Form.CompanyNumber
				lpa.TrustCorporation.Email = data.Form.Email

				// TODO: figure out what happens here
				lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.TrustCorporation, lpa.Attorneys, lpa.AttorneyDecisions)
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, appData.Paths.EnterTrustCorporationAddress.Format(lpa.ID))
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
		validation.Empty(),
		validation.Email())

	return errors
}
