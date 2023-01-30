package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App    AppData
	Errors validation.List
	Want   string
	Lpa    *Lpa
}

func WantReplacementAttorneys(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &wantReplacementAttorneysData{
			App:  appData,
			Want: lpa.WantReplacementAttorneys,
			Lpa:  lpa,
		}

		if r.Method == http.MethodPost {
			form := readWantReplacementAttorneysForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.WantReplacementAttorneys = form.Want
				var redirectUrl string

				if form.Want == "no" {
					lpa.ReplacementAttorneys = []Attorney{}
					lpa.Tasks.ChooseReplacementAttorneys = TaskCompleted
					redirectUrl = appData.Paths.TaskList
				} else {
					lpa.Tasks.ChooseReplacementAttorneys = TaskInProgress
					redirectUrl = appData.Paths.ChooseReplacementAttorneys
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		if len(lpa.ReplacementAttorneys) > 0 {
			return appData.Redirect(w, r, lpa, Paths.ChooseReplacementAttorneysSummary)
		}

		return tmpl(w, data)
	}
}

type wantReplacementAttorneysForm struct {
	Want string
}

func readWantReplacementAttorneysForm(r *http.Request) *wantReplacementAttorneysForm {
	return &wantReplacementAttorneysForm{
		Want: postFormString(r, "want"),
	}
}

func (f *wantReplacementAttorneysForm) Validate() validation.List {
	var errors validation.List

	if f.Want != "yes" && f.Want != "no" {
		errors.Add("want", "selectWantReplacementAttorneys")
	}

	return errors
}
