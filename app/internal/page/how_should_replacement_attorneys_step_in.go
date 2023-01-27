package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App    AppData
	Errors validation.List
	Form   *howShouldReplacementAttorneysStepInForm
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
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

			if data.Errors.Empty() {
				lpa.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if data.Form.WhenToStepIn != SomeOtherWay {
					lpa.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					lpa.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				redirectUrl := appData.Paths.TaskList

				if len(lpa.Attorneys) > 1 &&
					lpa.HowAttorneysMakeDecisions == JointlyAndSeverally &&
					lpa.HowShouldReplacementAttorneysStepIn == AllCanNoLongerAct &&
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
		WhenToStepIn: postFormString(r, "when-to-step-in"),
		OtherDetails: postFormString(r, "other-details"),
	}
}

func (f *howShouldReplacementAttorneysStepInForm) Validate() validation.List {
	var errors validation.List

	if f.WhenToStepIn == "" {
		errors.Add("when-to-step-in", "selectWhenToStepIn")
	}

	if f.WhenToStepIn == SomeOtherWay && f.OtherDetails == "" {
		errors.Add("other-details", "provideDetailsOfWhenToStepIn")
	}

	return errors
}
