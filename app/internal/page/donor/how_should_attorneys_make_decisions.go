package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howShouldAttorneysMakeDecisionsData struct {
	App    page.AppData
	Errors validation.List
	Form   *howShouldAttorneysMakeDecisionsForm
	Lpa    *page.Lpa
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.AttorneyDecisions.How,
				DecisionsDetails: lpa.AttorneyDecisions.Details,
			},
			Lpa: lpa,
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

				if lpa.AttorneyDecisions.RequiresHappiness(len(lpa.Attorneys)) {
					return appData.Redirect(w, r, lpa, page.Paths.AreYouHappyIfOneAttorneyCantActNoneCan.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
				}
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
		validation.Select(actor.Jointly, actor.JointlyAndSeverally, actor.JointlyForSomeSeverallyForOthers))

	if f.DecisionsType == actor.JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", "details", f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
