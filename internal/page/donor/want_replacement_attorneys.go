package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Options form.YesNoOptions
	Lpa     *actor.Lpa
}

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
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
				var redirectUrl page.LpaPath

				if lpa.WantReplacementAttorneys == form.No {
					lpa.ReplacementAttorneys = actor.Attorneys{}
					redirectUrl = page.Paths.TaskList
				} else {
					redirectUrl = page.Paths.ChooseReplacementAttorneys
				}

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return redirectUrl.Redirect(w, r, appData, lpa)
			}
		}

		if lpa.ReplacementAttorneys.Len() > 0 {
			return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, lpa)
		}

		return tmpl(w, data)
	}
}
