package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App               page.AppData
	Errors            validation.List
	AllowSomeOtherWay bool
	Form              *howShouldReplacementAttorneysStepInForm
	Options           page.ReplacementAttorneysStepInOptions
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldReplacementAttorneysStepInData{
			App:               appData,
			AllowSomeOtherWay: lpa.ReplacementAttorneys.Len() == 1,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: lpa.HowShouldReplacementAttorneysStepIn,
				OtherDetails: lpa.HowShouldReplacementAttorneysStepInDetails,
			},
			Options: page.ReplacementAttorneysStepInValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if lpa.HowShouldReplacementAttorneysStepIn != page.ReplacementAttorneysStepInAnotherWay {
					lpa.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					lpa.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.ReplacementAttorneys.Len() > 1 && lpa.HowShouldReplacementAttorneysStepIn == page.ReplacementAttorneysStepInWhenAllCanNoLongerAct {
					return appData.Redirect(w, r, lpa, appData.Paths.HowShouldReplacementAttorneysMakeDecisions.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, appData.Paths.TaskList.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}

type howShouldReplacementAttorneysStepInForm struct {
	WhenToStepIn page.ReplacementAttorneysStepIn
	Error        error
	OtherDetails string
}

func readHowShouldReplacementAttorneysStepInForm(r *http.Request) *howShouldReplacementAttorneysStepInForm {
	when, err := page.ParseReplacementAttorneysStepIn(page.PostFormString(r, "when-to-step-in"))

	return &howShouldReplacementAttorneysStepInForm{
		WhenToStepIn: when,
		Error:        err,
		OtherDetails: page.PostFormString(r, "other-details"),
	}
}

func (f *howShouldReplacementAttorneysStepInForm) Validate() validation.List {
	var errors validation.List

	errors.Error("when-to-step-in", "whenYourReplacementAttorneysStepIn", f.Error,
		validation.Selected())

	if f.WhenToStepIn == page.ReplacementAttorneysStepInAnotherWay {
		errors.String("other-details", "detailsOfWhenToStepIn", f.OtherDetails,
			validation.Empty())
	}

	return errors
}
