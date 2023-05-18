package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfRemainingReplacementAttorneysCanContinueToActData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfRemainingReplacementAttorneysCanContinueToActData{
			App:   appData,
			Happy: lpa.ReplacementAttorneyDecisions.HappyIfRemainingCanContinueToAct,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfRemainingReplacementAttorneysCanContinueToAct")

			if data.Errors.None() {
				lpa.ReplacementAttorneyDecisions.HappyIfRemainingCanContinueToAct = form.Happy
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.Type == page.LpaTypeHealthWelfare {
					return appData.Redirect(w, r, lpa, page.Paths.LifeSustainingTreatment)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.WhenCanTheLpaBeUsed)
				}
			}
		}

		return tmpl(w, data)
	}
}
