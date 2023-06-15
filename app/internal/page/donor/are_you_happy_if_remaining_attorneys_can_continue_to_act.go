package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type areYouHappyIfRemainingAttorneysCanContinueToActData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfRemainingAttorneysCanContinueToAct(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &areYouHappyIfRemainingAttorneysCanContinueToActData{
			App:   appData,
			Happy: lpa.AttorneyDecisions.HappyIfRemainingCanContinueToAct,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfRemainingAttorneysCanContinueToAct")

			if data.Errors.None() {
				lpa.AttorneyDecisions.HappyIfRemainingCanContinueToAct = form.Happy
				lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}
