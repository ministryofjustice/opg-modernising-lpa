package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &wantReplacementAttorneysData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(provided.WantReplacementAttorneys),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddReplacementAttorneys")
			data.Errors = f.Validate()

			if data.Errors.None() {
				provided.WantReplacementAttorneys = f.YesNo

				if provided.WantReplacementAttorneys.IsNo() {
					provided.ReplacementAttorneys = donordata.Attorneys{}
				}

				provided.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.WantReplacementAttorneys.IsYes() {
					return donor.PathChooseReplacementAttorneys.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
				} else {
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}
		}

		if provided.ReplacementAttorneys.Len() > 0 {
			return donor.PathChooseReplacementAttorneysSummary.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
