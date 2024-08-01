package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.DonorProvidedDetails
}

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
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

				if donor.WantReplacementAttorneys.IsNo() {
					donor.ReplacementAttorneys = donordata.Attorneys{}
				}

				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.WantReplacementAttorneys.IsYes() {
					return page.Paths.ChooseReplacementAttorneys.RedirectQuery(w, r, appData, donor, url.Values{"id": {newUID().String()}})
				} else {
					return page.Paths.TaskList.Redirect(w, r, appData, donor)
				}
			}
		}

		if donor.ReplacementAttorneys.Len() > 0 {
			return page.Paths.ChooseReplacementAttorneysSummary.Redirect(w, r, appData, donor)
		}

		return tmpl(w, data)
	}
}
