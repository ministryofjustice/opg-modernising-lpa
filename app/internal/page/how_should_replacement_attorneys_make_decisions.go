package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howShouldReplacementAttorneysMakeDecisionsData struct {
	App    AppData
	Errors map[string]string
	Form   *howShouldAttorneysMakeDecisionsForm
}

func HowShouldReplacementAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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
			data.Form = readHowShouldAttorneysMakeDecisionsForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.HowReplacementAttorneysMakeDecisions = data.Form.DecisionsType

				if data.Form.DecisionsType != JointlyForSomeSeverallyForOthers {
					lpa.HowReplacementAttorneysMakeDecisionsDetails = ""
				} else {
					lpa.HowReplacementAttorneysMakeDecisionsDetails = data.Form.DecisionsDetails
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, appData.Paths.TaskList, http.StatusFound)
			}
		}

		return tmpl(w, data)
	}
}
