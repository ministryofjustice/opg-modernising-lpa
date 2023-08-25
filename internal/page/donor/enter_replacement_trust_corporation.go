package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReplacementTrustCorporationData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReplacementTrustCorporationForm
	LpaID  string
}

func EnterReplacementTrustCorporation(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		trustCorporation := lpa.ReplacementAttorneys.TrustCorporation

		data := &enterReplacementTrustCorporationData{
			App: appData,
			Form: &enterReplacementTrustCorporationForm{
				Name:          trustCorporation.Name,
				CompanyNumber: trustCorporation.CompanyNumber,
				Email:         trustCorporation.Email,
			},
			LpaID: lpa.ID,
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReplacementTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
				trustCorporation.CompanyNumber = data.Form.CompanyNumber
				trustCorporation.Email = data.Form.Email
				lpa.ReplacementAttorneys.TrustCorporation = trustCorporation

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, appData.Paths.EnterReplacementTrustCorporationAddress.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type enterReplacementTrustCorporationForm struct {
	Name          string
	CompanyNumber string
	Email         string
}

func readEnterReplacementTrustCorporationForm(r *http.Request) *enterReplacementTrustCorporationForm {
	return &enterReplacementTrustCorporationForm{
		Name:          page.PostFormString(r, "name"),
		CompanyNumber: page.PostFormString(r, "company-number"),
		Email:         page.PostFormString(r, "email"),
	}
}

func (f *enterReplacementTrustCorporationForm) Validate() validation.List {
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
