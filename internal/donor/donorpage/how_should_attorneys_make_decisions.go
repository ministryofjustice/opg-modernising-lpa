package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
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
			App:     appData,
			Form:    newHowShouldAttorneysMakeDecisionsForm(appData.Localizer, "howAttorneysShouldMakeDecisions", "aRestrictionStatingWhichDecisionsYourAttorneysMustMakeJointly"),
			Donor:   provided,
			Options: lpadata.AttorneysActValues,
		}

		data.Form.DecisionsType.SetInput(provided.AttorneyDecisions.How)
		data.Form.DecisionsDetails.SetInput(provided.AttorneyDecisions.Details)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			provided.AttorneyDecisions.How = data.Form.DecisionsType.Value
			provided.AttorneyDecisions.Details = data.Form.DecisionsDetails.Value
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

		return tmpl(w, data)
	}
}

type howShouldAttorneysMakeDecisionsForm struct {
	newforms.Form
	DecisionsType    *newforms.Enum[lpadata.AttorneysAct, lpadata.AttorneysActOptions, *lpadata.AttorneysAct]
	DecisionsDetails *newforms.String
}

func newHowShouldAttorneysMakeDecisionsForm(l Localizer, errorLabel, detailsErrorLabel string) *howShouldAttorneysMakeDecisionsForm {
	return &howShouldAttorneysMakeDecisionsForm{
		DecisionsType: newforms.NewEnum[lpadata.AttorneysAct]("decision-type", errorLabel, lpadata.AttorneysActValues).
			Selected(),
		DecisionsDetails: newforms.NewString("mixed-details", detailsErrorLabel).
			NotEmpty().
			NoLinks(newforms.LocalizedError("yourInstructionsForAttorneys")),
	}
}

func (f *howShouldAttorneysMakeDecisionsForm) Parse(r *http.Request) bool {
	ok := f.ParsePostForm(r, f.DecisionsType)

	if f.DecisionsType.Value == lpadata.JointlyForSomeSeverallyForOthers {
		ok = f.ParsePostForm(r, f.DecisionsDetails) && ok
	}

	return ok
}
