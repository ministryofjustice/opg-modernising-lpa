package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type areYouHappyIfOneReplacementAttorneyCantActNoneCanData struct {
	App     page.AppData
	Errors  validation.List
	Happy   actor.YesNo
	Options actor.YesNoOptions
}

func AreYouHappyIfOneReplacementAttorneyCantActNoneCan(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App:     appData,
			Happy:   lpa.ReplacementAttorneyDecisions.HappyIfOneCannotActNoneCan,
			Options: actor.YesNoValues,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfOneReplacementAttorneyCantActNoneCan")

			if data.Errors.None() {
				lpa.ReplacementAttorneyDecisions.HappyIfOneCannotActNoneCan = form.Happy
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if form.Happy == actor.Yes {
					return appData.Redirect(w, r, lpa, page.Paths.TaskList.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}
