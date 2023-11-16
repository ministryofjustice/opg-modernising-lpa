package donor

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
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
		if lpa.Attorneys.Len() == 0 {
			return appData.Paths.ChooseAttorneys.RedirectQuery(w, r, appData, lpa, url.Values{"addAnother": {"1"}})
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
				redirectUrl := appData.Paths.TaskList
				if lpa.Attorneys.Len() > 1 {
					redirectUrl = appData.Paths.HowShouldAttorneysMakeDecisions
				}

				if data.Form.YesNo == form.Yes {
					return appData.Paths.ChooseAttorneys.RedirectQuery(w, r, appData, lpa, url.Values{"addAnother": {"1"}})
				} else {
					return redirectUrl.Redirect(w, r, appData, lpa)
				}
			}
		}

		return tmpl(w, data)
	}
}
