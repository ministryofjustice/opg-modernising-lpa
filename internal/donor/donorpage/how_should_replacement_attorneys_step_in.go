package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *howShouldReplacementAttorneysStepInForm
	Options lpadata.ReplacementAttorneysStepInOptions
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: donor.HowShouldReplacementAttorneysStepIn,
				OtherDetails: donor.HowShouldReplacementAttorneysStepInDetails,
			},
			Options: lpadata.ReplacementAttorneysStepInValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if donor.HowShouldReplacementAttorneysStepIn != lpadata.ReplacementAttorneysStepInAnotherWay {
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
	WhenToStepIn lpadata.ReplacementAttorneysStepIn
	Error        error
	OtherDetails string
}

func readHowShouldReplacementAttorneysStepInForm(r *http.Request) *howShouldReplacementAttorneysStepInForm {
	when, err := lpadata.ParseReplacementAttorneysStepIn(page.PostFormString(r, "when-to-step-in"))

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

	if f.WhenToStepIn == lpadata.ReplacementAttorneysStepInAnotherWay {
		errors.String("other-details", "detailsOfWhenToStepIn", f.OtherDetails,
			validation.Empty())
	}

	return errors
}
