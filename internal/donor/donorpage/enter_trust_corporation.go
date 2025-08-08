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
		trustCorporation, otherTrustCorporation := provided.Attorneys.TrustCorporation, provided.ReplacementAttorneys.TrustCorporation
		if service.IsReplacement() {
			trustCorporation, otherTrustCorporation = otherTrustCorporation, trustCorporation
		}

		data := &enterTrustCorporationData{
			App: appData,
			Form: &enterTrustCorporationForm{
				Name:  trustCorporation.Name,
				Email: trustCorporation.Email,
			},
			LpaID:               provided.LpaID,
			ChooseAttorneysPath: enterPath.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterTrustCorporationForm(r)
			data.Errors = data.Form.Validate(service.IsReplacement(), otherTrustCorporation)

			if data.Errors.None() {
				trustCorporation.Name = data.Form.Name
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
	Name  string
	Email string
}

func readEnterTrustCorporationForm(r *http.Request) *enterTrustCorporationForm {
	return &enterTrustCorporationForm{
		Name:  page.PostFormString(r, "name"),
		Email: page.PostFormString(r, "email"),
	}
}

func (f *enterTrustCorporationForm) Validate(isReplacement bool, otherTrustCorporation donordata.TrustCorporation) validation.List {
	var errors validation.List

	errors.String("name", "companyName", f.Name,
		validation.Empty())

	if f.Name == otherTrustCorporation.Name {
		errors.Add("name", trustCorporationCannotAlsoBeError{Name: f.Name, Replacement: isReplacement})

	}

	errors.String("email", "companyEmailAddress", f.Email,
		validation.Email())

	return errors
}

type trustCorporationCannotAlsoBeError struct {
	Name        string
	Replacement bool
}

func (e trustCorporationCannotAlsoBeError) Format(l validation.Localizer) string {
	isAppointed, cannotBe := "aReplacementAttorney", "anOriginalAttorney"
	if e.Replacement {
		isAppointed, cannotBe = cannotBe, isAppointed
	}

	return l.Format("errorTrustCorporationCannotAlsoBe", map[string]any{
		"Name":        e.Name,
		"IsAppointed": l.T(isAppointed),
		"CannotBe":    l.T(cannotBe),
	})
}
