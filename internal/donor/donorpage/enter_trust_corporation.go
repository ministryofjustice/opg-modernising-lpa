package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterTrustCorporationData struct {
	App                 appcontext.Data
	Errors              validation.List
	Form                *enterTrustCorporationForm
	LpaID               string
	ChooseAttorneysPath string
}

func EnterTrustCorporation(tmpl template.Template, service AttorneyService, newUID func() actoruid.UID) Handler {
	enterPath := donor.PathEnterAttorney
	addressPath := donor.PathEnterTrustCorporationAddress
	if service.IsReplacement() {
		enterPath = donor.PathEnterReplacementAttorney
		addressPath = donor.PathEnterReplacementTrustCorporationAddress
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		trustCorporation := provided.Attorneys.TrustCorporation
		if service.IsReplacement() {
			trustCorporation = provided.ReplacementAttorneys.TrustCorporation
		}

		data := &enterTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:          trustCorporation.Name,
				CompanyNumber: trustCorporation.CompanyNumber,
				Email:         trustCorporation.Email,
			},
			LpaID:               provided.LpaID,
			ChooseAttorneysPath: enterPath.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
				trustCorporation.CompanyNumber = data.Form.CompanyNumber
				trustCorporation.Email = data.Form.Email

				if err := service.PutTrustCorporation(r.Context(), provided, trustCorporation); err != nil {
					return err
				}

				return addressPath.Redirect(w, r, appData, provided)
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
