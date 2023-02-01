package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App    AppData
	Errors validation.List
	Form   *howShouldAttorneysMakeDecisionsForm
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howShouldReplacementAttorneysMakeDecisionsData{
			App: appData,
			Form: &howShouldAttorneysMakeDecisionsForm{
				DecisionsType:    lpa.HowReplacementAttorneysMakeDecisions,
				DecisionsDetails: lpa.HowReplacementAttorneysMakeDecisionsDetails,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r, "howReplacementAttorneysShouldMakeDecisions")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.HowReplacementAttorneysMakeDecisions = data.Form.DecisionsType

				if data.Form.DecisionsType != JointlyForSomeSeverallyForOthers {
					lpa.HowReplacementAttorneysMakeDecisionsDetails = ""
				} else {
					lpa.HowReplacementAttorneysMakeDecisionsDetails = data.Form.DecisionsDetails
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}
