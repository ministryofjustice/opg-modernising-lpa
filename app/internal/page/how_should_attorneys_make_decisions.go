package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldAttorneysMakeDecisionsData struct {
	App    AppData
	Errors validation.List
	Form   *howShouldAttorneysMakeDecisionsForm
	Lpa    *Lpa
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.HowAttorneysMakeDecisions,
				DecisionsDetails: lpa.HowAttorneysMakeDecisionsDetails,
			},
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howAttorneysShouldMakeDecisions")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.HowAttorneysMakeDecisions = data.Form.DecisionsType

				if data.Form.DecisionsType != JointlyForSomeSeverallyForOthers {
					lpa.HowAttorneysMakeDecisionsDetails = ""
				} else {
					lpa.HowAttorneysMakeDecisionsDetails = data.Form.DecisionsDetails
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.DoYouWantReplacementAttorneys)
			}
		}

		return tmpl(w, data)
	}
}

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType    string
	DecisionsDetails string
	errorLabel       string
}

func readHowShouldAttorneysMakeDecisionsForm(r *http.Request, errorLabel string) *howShouldAttorneysMakeDecisionsForm {
	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType:    postFormString(r, "decision-type"),
		DecisionsDetails: postFormString(r, "mixed-details"),
		errorLabel:       errorLabel,
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	errors.String("decision-type", f.errorLabel, f.DecisionsType,
		validation.Select(Jointly, JointlyAndSeverally, JointlyForSomeSeverallyForOthers))

	if f.DecisionsType == JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", "details", f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
