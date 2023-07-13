package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type wantReplacementAttorneysData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Options form.YesNoOptions
	Lpa     *page.Lpa
}

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &wantReplacementAttorneysData{
			App: appData,
			Lpa: lpa,
			Form: &form.YesNoForm{
				YesNo: lpa.WantReplacementAttorneys,
			},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddReplacementAttorneys")
			data.Errors = f.Validate()

			if data.Errors.None() {
				lpa.WantReplacementAttorneys = f.YesNo
				var redirectUrl string

				if lpa.WantReplacementAttorneys == form.No {
					lpa.ReplacementAttorneys = actor.Attorneys{}
					redirectUrl = page.Paths.TaskList.Format(lpa.ID)
				} else {
					redirectUrl = page.Paths.ChooseReplacementAttorneys.Format(lpa.ID)
				}

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		if len(lpa.ReplacementAttorneys) > 0 {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary.Format(lpa.ID))
		}

		return tmpl(w, data)
	}
}
