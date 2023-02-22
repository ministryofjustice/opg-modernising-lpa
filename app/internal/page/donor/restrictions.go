package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type restrictionsData struct {
	App       page.AppData
	Errors    validation.List
	Completed bool
	Lpa       *page.Lpa
}

func Restrictions(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
					lpa.Tasks.Restrictions = page.TaskInProgress
				} else {
					lpa.Tasks.Restrictions = page.TaskCompleted
					lpa.Restrictions = form.Restrictions
				}
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.WhoDoYouWantToBeCertificateProviderGuidance)
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
		AnswerLater:  page.PostFormString(r, "answer-later") == "1",
		Restrictions: page.PostFormString(r, "restrictions"),
	}
}

func (f *restrictionsForm) Validate() validation.List {
	var errors validation.List

	errors.String("restrictions", "restrictions", f.Restrictions,
		validation.StringTooLong(10000))

	return errors
}
