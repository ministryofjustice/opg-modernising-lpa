package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type restrictionsData struct {
	App       AppData
	Errors    validation.List
	Completed bool
	Lpa       *Lpa
}

func Restrictions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &restrictionsData{
			App:       appData,
			Completed: lpa.Tasks.Restrictions.Completed(),
			Lpa:       lpa,
		}

		if r.Method == http.MethodPost {
			form := readRestrictionsForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() || form.AnswerLater {
				if form.AnswerLater {
					lpa.Tasks.Restrictions = TaskInProgress
				} else {
					lpa.Tasks.Restrictions = TaskCompleted
					lpa.Restrictions = form.Restrictions
				}
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.WhoDoYouWantToBeCertificateProviderGuidance)
			}
		}

		return tmpl(w, data)
	}
}

type restrictionsForm struct {
	AnswerLater  bool
	Restrictions string
}

func readRestrictionsForm(r *http.Request) *restrictionsForm {
	return &restrictionsForm{
		AnswerLater:  postFormString(r, "answer-later") == "1",
		Restrictions: postFormString(r, "restrictions"),
	}
}

func (f *restrictionsForm) Validate() validation.List {
	var errors validation.List

	if len(f.Restrictions) > 10000 {
		errors.Add("restrictions", "restrictionsTooLong")
	}

	return errors
}
