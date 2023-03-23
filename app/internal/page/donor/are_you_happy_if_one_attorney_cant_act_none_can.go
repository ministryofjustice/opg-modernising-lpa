package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfOneAttorneyCantActNoneCanData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfOneAttorneyCantActNoneCan(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfOneAttorneyCantActNoneCanData{
			App:   appData,
			Happy: lpa.HowAttorneysMakeDecisions.HappyIfOneCannotActNoneCan,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfOneAttorneyCantActNoneCan")

			if data.Errors.None() {
				lpa.HowAttorneysMakeDecisions.HappyIfOneCannotActNoneCan = form.Happy

				redirect := page.Paths.AreYouHappyIfRemainingAttorneysCanContinueToAct
				if form.Happy == "yes" {
					redirect = page.Paths.DoYouWantReplacementAttorneys
					lpa.Tasks.ChooseAttorneys = page.TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirect)
			}
		}

		return tmpl(w, data)
	}
}
