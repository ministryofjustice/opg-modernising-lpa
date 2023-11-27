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
	Donor   *actor.DonorProvidedDetails
	Options actor.AttorneysActOptions
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    donor.AttorneyDecisions.How,
				DecisionsDetails: donor.AttorneyDecisions.Details,
			},
			Donor:   donor,
			Options: actor.AttorneysActValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howAttorneysShouldMakeDecisions", "detailsAboutTheDecisionsYourAttorneysMustMakeTogether")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.AttorneyDecisions = actor.MakeAttorneyDecisions(
					donor.AttorneyDecisions,
					data.Form.DecisionsType,
					data.Form.DecisionsDetails)
				donor.Tasks.ChooseAttorneys = page.ChooseAttorneysState(donor.Attorneys, donor.AttorneyDecisions)
				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType     actor.AttorneysAct
	Error             error
	DecisionsDetails  string
	errorLabel        string
	detailsErrorLabel string
}

func readHowShouldAttorneysMakeDecisionsForm(r *http.Request, errorLabel, detailsErrorLabel string) *howShouldAttorneysMakeDecisionsForm {
	how, err := actor.ParseAttorneysAct(page.PostFormString(r, "decision-type"))

	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType:     how,
		Error:             err,
		DecisionsDetails:  page.PostFormString(r, "mixed-details"),
		errorLabel:        errorLabel,
		detailsErrorLabel: detailsErrorLabel,
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	errors.Error("decision-type", f.errorLabel, f.Error,
		validation.Selected())

	if f.DecisionsType == actor.JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", f.detailsErrorLabel, f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
