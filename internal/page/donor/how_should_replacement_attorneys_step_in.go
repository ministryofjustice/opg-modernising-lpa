package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App     page.AppData
	Errors  validation.List
	Form    *howShouldReplacementAttorneysStepInForm
	Options donordata.ReplacementAttorneysStepInOptions
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: donor.HowShouldReplacementAttorneysStepIn,
				OtherDetails: donor.HowShouldReplacementAttorneysStepInDetails,
			},
			Options: donordata.ReplacementAttorneysStepInValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if donor.HowShouldReplacementAttorneysStepIn != actor.ReplacementAttorneysStepInAnotherWay {
					donor.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					donor.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				donor.Tasks.ChooseReplacementAttorneys = page.ChooseReplacementAttorneysState(donor)

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.ReplacementAttorneys.Len() > 1 && donor.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() {
					return page.Paths.HowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.TaskList.Redirect(w, r, appData, donor)
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
	when, err := donordata.ParseReplacementAttorneysStepIn(page.PostFormString(r, "when-to-step-in"))

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
