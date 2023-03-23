package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfRemainingAttorneysCanContinueToActData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfRemainingAttorneysCanContinueToAct(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfRemainingAttorneysCanContinueToActData{
			App:   appData,
			Happy: lpa.AttorneyDecisions.HappyIfRemainingCanContinueToAct,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfRemainingAttorneysCanContinueToAct")

			if data.Errors.None() {
				lpa.AttorneyDecisions.HappyIfRemainingCanContinueToAct = form.Happy
				lpa.Tasks.ChooseAttorneys = page.TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys)
			}
		}

		return tmpl(w, data)
	}
}
