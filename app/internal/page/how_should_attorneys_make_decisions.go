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

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType    string
	DecisionsDetails string
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
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.Empty() {
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

func readHowShouldAttorneysMakeDecisionsForm(r *http.Request) *howShouldAttorneysMakeDecisionsForm {
	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType:    postFormString(r, "decision-type"),
		DecisionsDetails: postFormString(r, "mixed-details"),
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	if f.DecisionsType != JointlyAndSeverally && f.DecisionsType != Jointly && f.DecisionsType != JointlyForSomeSeverallyForOthers {
		errors.Add("decision-type", "chooseADecisionType")
	}

	if f.DecisionsType == JointlyForSomeSeverallyForOthers && f.DecisionsDetails == "" {
		errors.Add("mixed-details", "provideDecisionDetails")
	}

	return errors
}
