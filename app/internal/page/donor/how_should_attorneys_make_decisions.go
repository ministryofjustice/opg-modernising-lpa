package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldAttorneysMakeDecisionsData struct {
	App    page.AppData
	Errors validation.List
	Form   *howShouldAttorneysMakeDecisionsForm
	Lpa    *page.Lpa
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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

				if data.Form.DecisionsType != page.JointlyForSomeSeverallyForOthers {
					lpa.HowAttorneysMakeDecisionsDetails = ""
				} else {
					lpa.HowAttorneysMakeDecisionsDetails = data.Form.DecisionsDetails
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys)
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
		DecisionsType:    page.PostFormString(r, "decision-type"),
		DecisionsDetails: page.PostFormString(r, "mixed-details"),
		errorLabel:       errorLabel,
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	errors.String("decision-type", f.errorLabel, f.DecisionsType,
		validation.Select(page.Jointly, page.JointlyAndSeverally, page.JointlyForSomeSeverallyForOthers))

	if f.DecisionsType == page.JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", "details", f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
