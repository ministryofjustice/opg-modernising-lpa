package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signData struct {
	App                        page.AppData
	Errors                     validation.List
	Attorney                   actor.Attorney
	IsReplacement              bool
	LpaCanBeUsedWhenRegistered bool
	Form                       *signForm
}

func Sign(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney {
			attorneys = lpa.ReplacementAttorneys
		}

		attorney, ok := attorneys.Get(appData.AttorneyID)
		if !ok {
			return appData.Redirect(w, r, lpa, page.Paths.Attorney.Start)
		}

		attorneyProvidedDetails := getProvidedDetails(appData, lpa)

		data := &signData{
			App:                        appData,
			Attorney:                   attorney,
			IsReplacement:              appData.IsReplacementAttorney,
			LpaCanBeUsedWhenRegistered: lpa.WhenCanTheLpaBeUsed == page.UsedWhenRegistered,
			Form: &signForm{
				Confirm: attorneyProvidedDetails.Confirmed,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSignForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.Confirmed = true
				if appData.IsReplacementAttorney {
					lpa.ReplacementAttorneyProvidedDetails.Put(attorneyProvidedDetails)
				} else {
					lpa.AttorneyProvidedDetails.Put(attorneyProvidedDetails)
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.Attorney.NextPage)
			}
		}

		return tmpl(w, data)
	}
}

type signForm struct {
	Confirm bool
}

func readSignForm(r *http.Request) *signForm {
	return &signForm{
		Confirm: page.PostFormString(r, "confirm") == "1",
	}
}

func (f *signForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("confirm", "confirm", f.Confirm,
		validation.Selected())

	return errors
}
