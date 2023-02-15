package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wantReplacementAttorneysData struct {
	App    page.AppData
	Errors validation.List
	Want   string
	Lpa    *page.Lpa
}

func WantReplacementAttorneys(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
			f := readWantReplacementAttorneysForm(r)
			data.Errors = f.Validate()

			if data.Errors.None() {
				lpa.WantReplacementAttorneys = f.Want
				var redirectUrl string

				if f.Want == "no" {
					lpa.ReplacementAttorneys = actor.Attorneys{}
					lpa.Tasks.ChooseReplacementAttorneys = page.TaskCompleted
					redirectUrl = appData.Paths.TaskList
				} else {
					lpa.Tasks.ChooseReplacementAttorneys = page.TaskInProgress
					redirectUrl = appData.Paths.ChooseReplacementAttorneys
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		if len(lpa.ReplacementAttorneys) > 0 {
			return appData.Redirect(w, r, lpa, page.Paths.ChooseReplacementAttorneysSummary)
		}

		return tmpl(w, data)
	}
}

type wantReplacementAttorneysForm struct {
	Want string
}

func readWantReplacementAttorneysForm(r *http.Request) *wantReplacementAttorneysForm {
	return &wantReplacementAttorneysForm{
		Want: page.PostFormString(r, "want"),
	}
}

func (f *wantReplacementAttorneysForm) Validate() validation.List {
	var errors validation.List

	errors.String("want", "yesToAddReplacementAttorneys", f.Want,
		validation.Select("yes", "no"))

	return errors
}
