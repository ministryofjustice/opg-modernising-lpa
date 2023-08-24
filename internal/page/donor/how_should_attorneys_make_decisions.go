package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldAttorneysMakeDecisionsData struct {
	App     page.AppData
	Errors  validation.List
	Form    *howShouldAttorneysMakeDecisionsForm
	Lpa     *page.Lpa
	Options actor.AttorneysActOptions
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.AttorneyDecisions.How,
				DecisionsDetails: lpa.AttorneyDecisions.Details,
			},
			Lpa:     lpa,
			Options: actor.AttorneysActValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howAttorneysShouldMakeDecisions")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.AttorneyDecisions = actor.MakeAttorneyDecisions(
					lpa.AttorneyDecisions,
					data.Form.DecisionsType,
					data.Form.DecisionsDetails)
				lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType    actor.AttorneysAct
	Error            error
	DecisionsDetails string
	errorLabel       string
}

func readHowShouldAttorneysMakeDecisionsForm(r *http.Request, errorLabel string) *howShouldAttorneysMakeDecisionsForm {
	how, err := actor.ParseAttorneysAct(page.PostFormString(r, "decision-type"))

	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType:    how,
		Error:            err,
		DecisionsDetails: page.PostFormString(r, "mixed-details"),
		errorLabel:       errorLabel,
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	errors.Error("decision-type", f.errorLabel, f.Error,
		validation.Selected())

	if f.DecisionsType == actor.JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", "details", f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
