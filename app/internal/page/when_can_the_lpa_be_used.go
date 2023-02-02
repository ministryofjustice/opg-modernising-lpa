package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whenCanTheLpaBeUsedData struct {
	App       AppData
	Errors    validation.List
	When      string
	Completed bool
	Lpa       *Lpa
}

func WhenCanTheLpaBeUsed(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whenCanTheLpaBeUsedData{
			App:       appData,
			When:      lpa.WhenCanTheLpaBeUsed,
			Completed: lpa.Tasks.WhenCanTheLpaBeUsed.Completed(),
			Lpa:       lpa,
		}

		if r.Method == http.MethodPost {
			form := readWhenCanTheLpaBeUsedForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() || form.AnswerLater {
				if form.AnswerLater {
					lpa.Tasks.WhenCanTheLpaBeUsed = TaskInProgress
				} else {
					lpa.WhenCanTheLpaBeUsed = form.When
					lpa.Tasks.WhenCanTheLpaBeUsed = TaskCompleted
				}
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.Restrictions)
			}
		}

		return tmpl(w, data)
	}
}

type whenCanTheLpaBeUsedForm struct {
	AnswerLater bool
	When        string
}

func readWhenCanTheLpaBeUsedForm(r *http.Request) *whenCanTheLpaBeUsedForm {
	return &whenCanTheLpaBeUsedForm{
		AnswerLater: postFormString(r, "answer-later") == "1",
		When:        postFormString(r, "when"),
	}
}

func (f *whenCanTheLpaBeUsedForm) Validate() validation.List {
	var errors validation.List

	errors.String("when", "whenYourAttorneysCanUseYourLpa", f.When,
		validation.Select(UsedWhenRegistered, UsedWhenCapacityLost))

	return errors
}
