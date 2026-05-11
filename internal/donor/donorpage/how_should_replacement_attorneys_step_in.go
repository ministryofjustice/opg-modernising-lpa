package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howShouldReplacementAttorneysStepInData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *howShouldReplacementAttorneysStepInForm
	Options lpadata.ReplacementAttorneysStepInOptions
}

func HowShouldReplacementAttorneysStepIn(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howShouldReplacementAttorneysStepInData{
			App: appData,
			Form: &howShouldReplacementAttorneysStepInForm{
				WhenToStepIn: provided.HowShouldReplacementAttorneysStepIn,
				OtherDetails: provided.HowShouldReplacementAttorneysStepInDetails,
			},
			Options: lpadata.ReplacementAttorneysStepInValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowShouldReplacementAttorneysStepInForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.HowShouldReplacementAttorneysStepIn = data.Form.WhenToStepIn

				if provided.HowShouldReplacementAttorneysStepIn != lpadata.ReplacementAttorneysStepInAnotherWay {
					provided.HowShouldReplacementAttorneysStepInDetails = ""
				} else {
					provided.HowShouldReplacementAttorneysStepInDetails = data.Form.OtherDetails
				}

				provided.Tasks.ChooseReplacementAttorneys = donordata.ChooseReplacementAttorneysState(provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.ReplacementAttorneys.Len() > 1 && provided.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() {
					return donor.PathHowShouldReplacementAttorneysMakeDecisions.Redirect(w, r, appData, provided)
				} else {
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}

type howShouldReplacementAttorneysStepInForm struct {
	newforms.Form
	WhenToStepIn *newforms.Enum[lpadata.ReplacementAttorneysStepIn, lpadata.ReplacementAttorneysStepInOptions, *lpadata.ReplacementAttorneysStepIn]
	OtherDetails *newforms.String
}

func newHowShouldReplacementAttorneysStepInForm(l Localizer) *howShouldReplacementAttorneysStepInForm {
	return &howShouldReplacementAttorneysStepInForm{
		WhenToStepIn: newforms.NewEnum[lpadata.ReplacementAttorneysStepIn]("when-to-step-in", l.T("whenYourReplacementAttorneysStepIn"), lpadata.ReplacementAttorneysStepInValues).
			Selected(),
		OtherDetails: newforms.NewString("other-details", l.T("detailsOfWhenToStepIn")).
			NotEmpty().
			NoLinks(newforms.LocalizedError("yourInstructionsForAttorneys")),
	}
}

func (f *howShouldReplacementAttorneysStepInForm) Parse(r *http.Request) bool {
	ok := f.ParsePostForm(r, f.WhenToStepIn)

	if f.WhenToStepIn.Value == lpadata.ReplacementAttorneysStepInAnotherWay {
		ok = f.ParsePostForm(r, f.OtherDetails) && ok
	}

	return ok
}
