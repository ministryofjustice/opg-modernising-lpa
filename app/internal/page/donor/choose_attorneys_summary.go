package donor

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type chooseAttorneysSummaryData struct {
	App     page.AppData
	Errors  validation.List
	Form    *form.YesNoForm
	Lpa     *page.Lpa
	Options form.YesNoOptions
}

func ChooseAttorneysSummary(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if len(lpa.Attorneys) == 0 {
			return appData.Redirect(w, r, lpa, fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseAttorneys.Format(lpa.ID)))
		}

		data := &chooseAttorneysSummaryData{
			App:     appData,
			Lpa:     lpa,
			Form:    &form.YesNoForm{},
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := appData.Paths.TaskList.Format(lpa.ID)

				if len(lpa.Attorneys) > 1 {
					redirectUrl = appData.Paths.HowShouldAttorneysMakeDecisions.Format(lpa.ID)
				}

				if data.Form.YesNo == form.Yes {
					redirectUrl = fmt.Sprintf("%s?addAnother=1", appData.Paths.ChooseAttorneys.Format(lpa.ID))
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}
