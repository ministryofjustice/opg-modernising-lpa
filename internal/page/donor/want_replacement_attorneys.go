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
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *actor.DonorProvidedDetails
}

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &wantReplacementAttorneysData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(donor.WantReplacementAttorneys),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddReplacementAttorneys")
			data.Errors = f.Validate()

			if data.Errors.None() {
				donor.WantReplacementAttorneys = f.YesNo
				var redirectUrl page.LpaPath

				if donor.WantReplacementAttorneys == form.No {
					donor.ReplacementAttorneys = actor.Attorneys{}
					redirectUrl = page.Paths.TaskList
				} else {
					redirectUrl = page.Paths.ChooseReplacementAttorneys
				}

				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirectUrl.Redirect(w, r, appData, donor)
			}
		}

		if donor.ReplacementAttorneys.Len() > 0 {
			return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, donor)
		}

		return tmpl(w, data)
	}
}
