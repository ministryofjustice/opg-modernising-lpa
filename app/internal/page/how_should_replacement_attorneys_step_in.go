package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howShouldReplacementAttorneysStepInData struct {
	App    AppData
	Errors map[string]string
	Form   *howShouldReplacementAttorneysStepInForm
}

type howShouldReplacementAttorneysStepInForm struct {
	WhenToStepIn string
	OtherDetails string
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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

			if len(data.Errors) == 0 {
				lpa.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if data.Form.WhenToStepIn != SomeOtherWay {
					lpa.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					lpa.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				redirectUrl := taskListPath

				if len(lpa.Attorneys) > 1 &&
					lpa.HowAttorneysMakeDecisions == JointlyAndSeverally &&
					lpa.HowShouldReplacementAttorneysStepIn == AllCanNoLongerAct &&
					len(lpa.ReplacementAttorneys) > 1 {
					redirectUrl = howShouldReplacementAttorneysMakeDecisionsPath

				}

				appData.Lang.Redirect(w, r, redirectUrl, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

func readHowShouldReplacementAttorneysStepInForm(r *http.Request) *howShouldReplacementAttorneysStepInForm {
	return &howShouldReplacementAttorneysStepInForm{
		WhenToStepIn: postFormString(r, "when-to-step-in"),
		OtherDetails: postFormString(r, "other-details"),
	}
}

func (f *howShouldReplacementAttorneysStepInForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.WhenToStepIn == "" {
		errors["when-to-step-in"] = "selectWhenToStepIn"
	}

	if f.WhenToStepIn == SomeOtherWay && f.OtherDetails == "" {
		errors["other-details"] = "provideDetailsOfWhenToStepIn"
	}

	return errors
}
