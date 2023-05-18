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

func WantReplacementAttorneys(tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
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
					lpa.ReplacementAttorneys = actor.Attorneys{}
					if lpa.Type == page.LpaTypeHealthWelfare {
						redirectUrl = page.Paths.LifeSustainingTreatment
					} else {
						redirectUrl = page.Paths.WhenCanTheLpaBeUsed
					}
				} else {
					redirectUrl = appData.Paths.ChooseReplacementAttorneys
				}

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
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
