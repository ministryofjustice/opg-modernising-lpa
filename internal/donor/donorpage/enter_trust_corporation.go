package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
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
			App:                 appData,
			Form:                newEnterTrustCorporationForm(appData.Localizer),
			LpaID:               provided.LpaID,
			ChooseAttorneysPath: enterPath.FormatQuery(provided.LpaID, url.Values{"id": {newUID().String()}}),
		}

		data.Form.Name.SetInput(trustCorporation.Name)
		data.Form.Email.SetInput(trustCorporation.Email)

		if r.Method == http.MethodPost && data.Form.Parse(r, service.IsReplacement(), otherTrustCorporation) {
			trustCorporation.Name = data.Form.Name.Value
			trustCorporation.Email = data.Form.Email.Value

			if err := service.PutTrustCorporation(r.Context(), provided, trustCorporation); err != nil {
				return err
			}

			return addressPath.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}

type enterTrustCorporationForm struct {
	newforms.Form
	Name  *newforms.String
	Email *newforms.String
}

func newEnterTrustCorporationForm(l Localizer) *enterTrustCorporationForm {
	return &enterTrustCorporationForm{
		Name: newforms.NewString("name", l.T("trustCorporationName")).
			NotEmpty(),
		Email: newforms.NewString("email", l.T("trustCorporationEmailAddress")).
			Email(),
	}
}

func (f *enterTrustCorporationForm) Parse(r *http.Request, isReplacement bool, otherTrustCorporation donordata.TrustCorporation) bool {
	ok := f.ParsePostForm(r, f.Name, f.Email)

	if f.Name.Value == otherTrustCorporation.Name {
		f.Name.Error = trustCorporationCannotAlsoBeError{Name: f.Name.Value, Replacement: isReplacement}
		f.Errors = append(f.Errors, f.Name.Field)
		ok = false
	}

	return ok
}

type trustCorporationCannotAlsoBeError struct {
	Name        string
	Replacement bool
}

func (e trustCorporationCannotAlsoBeError) Format(l newforms.Localizer) string {
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
