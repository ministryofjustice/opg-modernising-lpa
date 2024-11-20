package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldAttorneysMakeDecisionsData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *howShouldAttorneysMakeDecisionsForm
	Donor   *donordata.Provided
	Options lpadata.AttorneysActOptions
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howShouldAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    provided.AttorneyDecisions.How,
				DecisionsDetails: provided.AttorneyDecisions.Details,
			},
			Donor:   provided,
			Options: lpadata.AttorneysActValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howAttorneysShouldMakeDecisions", "detailsAboutTheDecisionsYourAttorneysMustMakeJointly")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.AttorneyDecisions.How = data.Form.DecisionsType
				provided.AttorneyDecisions.Details = data.Form.DecisionsDetails
				provided.UpdateDecisions()
				provided.Tasks.ChooseAttorneys = donordata.ChooseAttorneysState(provided.Attorneys, provided.AttorneyDecisions)
				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				switch provided.AttorneyDecisions.How {
				case lpadata.Jointly:
					return donor.PathBecauseYouHaveChosenJointly.Redirect(w, r, appData, provided)
				case lpadata.JointlyForSomeSeverallyForOthers:
					return donor.PathBecauseYouHaveChosenJointlyForSomeSeverallyForOthers.Redirect(w, r, appData, provided)
				default:
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType     lpadata.AttorneysAct
	DecisionsDetails  string
	errorLabel        string
	detailsErrorLabel string
}

func readHowShouldAttorneysMakeDecisionsForm(r *http.Request, errorLabel, detailsErrorLabel string) *howShouldAttorneysMakeDecisionsForm {
	how, _ := lpadata.ParseAttorneysAct(page.PostFormString(r, "decision-type"))

	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType:     how,
		DecisionsDetails:  page.PostFormString(r, "mixed-details"),
		errorLabel:        errorLabel,
		detailsErrorLabel: detailsErrorLabel,
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("decision-type", f.errorLabel, f.DecisionsType,
		validation.Selected())

	if f.DecisionsType == lpadata.JointlyForSomeSeverallyForOthers {
		errors.String("mixed-details", f.detailsErrorLabel, f.DecisionsDetails,
			validation.Empty())
	}

	return errors
}
