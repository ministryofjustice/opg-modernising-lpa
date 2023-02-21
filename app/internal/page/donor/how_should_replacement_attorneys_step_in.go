package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App    page.AppData
	Errors validation.List
	Form   *howShouldReplacementAttorneysStepInForm
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: lpa.HowShouldReplacementAttorneysStepIn,
				OtherDetails: lpa.HowShouldReplacementAttorneysStepInDetails,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if data.Form.WhenToStepIn != page.SomeOtherWay {
					lpa.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					lpa.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				redirectUrl := appData.Paths.TaskList

				if len(lpa.Attorneys) > 1 &&
					lpa.HowAttorneysMakeDecisions == page.JointlyAndSeverally &&
					lpa.HowShouldReplacementAttorneysStepIn == page.AllCanNoLongerAct &&
					len(lpa.ReplacementAttorneys) > 1 {
					redirectUrl = appData.Paths.HowShouldReplacementAttorneysMakeDecisions

				}

				return appData.Redirect(w, r, lpa, redirectUrl)
			}
		}

		return tmpl(w, data)
	}
}

type howShouldReplacementAttorneysStepInForm struct {
	WhenToStepIn string
	OtherDetails string
}

func readHowShouldReplacementAttorneysStepInForm(r *http.Request) *howShouldReplacementAttorneysStepInForm {
	return &howShouldReplacementAttorneysStepInForm{
		WhenToStepIn: page.PostFormString(r, "when-to-step-in"),
		OtherDetails: page.PostFormString(r, "other-details"),
	}
}

func (f *howShouldReplacementAttorneysStepInForm) Validate() validation.List {
	var errors validation.List

	errors.String("when-to-step-in", "whenYourReplacementAttorneysStepIn", f.WhenToStepIn,
		validation.Select(page.OneCanNoLongerAct, page.AllCanNoLongerAct, page.SomeOtherWay))

	if f.WhenToStepIn == page.SomeOtherWay {
		errors.String("other-details", "detailsOfWhenToStepIn", f.OtherDetails,
			validation.Empty())
	}

	return errors
}
