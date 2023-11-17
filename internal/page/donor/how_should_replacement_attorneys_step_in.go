package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App     page.AppData
	Errors  validation.List
	Form    *howShouldReplacementAttorneysStepInForm
	Options actor.ReplacementAttorneysStepInOptions
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: lpa.HowShouldReplacementAttorneysStepIn,
				OtherDetails: lpa.HowShouldReplacementAttorneysStepInDetails,
			},
			Options: actor.ReplacementAttorneysStepInValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if lpa.HowShouldReplacementAttorneysStepIn != actor.ReplacementAttorneysStepInAnotherWay {
					lpa.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					lpa.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				lpa.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(lpa)

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.ReplacementAttorneys.Len() > 1 && lpa.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() {
					return appData.Paths.HowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, lpa)
				} else {
					return appData.Paths.TaskList.Redirect(w, r, appData, lpa)
				}
			}
		}

		return tmpl(w, data)
	}
}

type howShouldReplacementAttorneysStepInForm struct {
	WhenToStepIn actor.ReplacementAttorneysStepIn
	Error        error
	OtherDetails string
}

func readHowShouldReplacementAttorneysStepInForm(r *http.Request) *howShouldReplacementAttorneysStepInForm {
	when, err := actor.ParseReplacementAttorneysStepIn(page.PostFormString(r, "when-to-step-in"))

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

	if f.WhenToStepIn == actor.ReplacementAttorneysStepInAnotherWay {
		errors.String("other-details", "detailsOfWhenToStepIn", f.OtherDetails,
			validation.Empty())
	}

	return errors
}
