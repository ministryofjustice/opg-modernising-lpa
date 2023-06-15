package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type areYouHappyIfOneAttorneyCantActNoneCanData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfOneAttorneyCantActNoneCan(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &areYouHappyIfOneAttorneyCantActNoneCanData{
			App:   appData,
			Happy: lpa.AttorneyDecisions.HappyIfOneCannotActNoneCan,
		}

		if r.Method == http.MethodPost {
			form := form.ReadHappyForm(r)
			data.Errors = form.Validate("yesIfYouAreHappyIfOneAttorneyCantActNoneCan")

			if data.Errors.None() {
				lpa.AttorneyDecisions.HappyIfOneCannotActNoneCan = form.Happy
				lpa.Tasks.ChooseAttorneys = page.ChooseAttorneysState(lpa.Attorneys, lpa.AttorneyDecisions)
				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if form.Happy == "yes" {
					return appData.Redirect(w, r, lpa, page.Paths.TaskList)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.AreYouHappyIfRemainingAttorneysCanContinueToAct)
				}
			}
		}

		return tmpl(w, data)
	}
}
