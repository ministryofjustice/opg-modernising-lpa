package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func WantReplacementAttorneys(tmpl template.Template, service AttorneyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.ReplacementAttorneys.Len() > 0 {
			return donor.PathChooseReplacementAttorneysSummary.Redirect(w, r, appData, provided)
		}

		data := &wantReplacementAttorneysData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(provided.WantReplacementAttorneys),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddReplacementAttorneys")
			data.Errors = f.Validate()

			if data.Errors.None() {
				if err := service.WantReplacements(r.Context(), provided, f.YesNo); err != nil {
					return err
				}

				if f.YesNo.IsYes() {
					return donor.PathChooseReplacementAttorneys.Redirect(w, r, appData, provided)
				} else {
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
