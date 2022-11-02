package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howShouldAttorneysMakeDecisionsData struct {
	App              AppData
	DecisionsType    string
	DecisionsDetails string
	Errors           map[string]string
}

type howShouldAttorneysMakeDecisionsForm struct {
	DecisionsType    string
	DecisionsDetails string
}

func HowShouldAttorneysMakeDecisions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			fmt.Print(lpa.ID)
			return err
		}

		data := &howShouldAttorneysMakeDecisionsData{
			App:              appData,
			DecisionsType:    lpa.DecisionsType,
			DecisionsDetails: lpa.DecisionsDetails,
		}

		if r.Method == http.MethodPost {
			form := readHowShouldAttorneysMakeDecisionsForm(r)
			//data.Errors = form.Validate()

			//if len(data.Errors) == 0 {
			//	lpa.Tasks.CertificateProvider = TaskCompleted
			lpa.DecisionsType = form.DecisionsType
			lpa.DecisionsDetails = form.DecisionsDetails
			if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
				return err
			}
			appData.Lang.Redirect(w, r, wantReplacementAttorneysPath, http.StatusFound)
			return nil
			//}
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
