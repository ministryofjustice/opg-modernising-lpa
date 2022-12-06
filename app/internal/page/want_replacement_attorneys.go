package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type wantReplacementAttorneysData struct {
	App    AppData
	Errors map[string]string
	Want   string
	Lpa    *Lpa
}

func WantReplacementAttorneys(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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

			if len(data.Errors) == 0 {
				lpa.WantReplacementAttorneys = form.Want
				var redirectUrl string

				if form.Want == "no" {
					lpa.ReplacementAttorneys = []Attorney{}
					redirectUrl = appData.Paths.TaskList
				} else {
					redirectUrl = appData.Paths.ChooseReplacementAttorneys
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
				return nil
			}
		}

		if len(lpa.ReplacementAttorneys) > 0 {
			appData.Lang.Redirect(w, r, appData.Paths.ChooseReplacementAttorneysSummary, http.StatusFound)
			return nil
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

func (f *wantReplacementAttorneysForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.Want != "yes" && f.Want != "no" {
		errors["want"] = "selectWantReplacementAttorneys"
	}

	return errors
}
