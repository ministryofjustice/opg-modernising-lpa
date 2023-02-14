package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whenCanTheLpaBeUsedData struct {
	App       page.AppData
	Errors    validation.List
	When      string
	Completed bool
	Lpa       *page.Lpa
}

func WhenCanTheLpaBeUsed(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
			f := readWhenCanTheLpaBeUsedForm(r)
			data.Errors = f.Validate()

			if data.Errors.None() || f.AnswerLater {
				if f.AnswerLater {
					lpa.Tasks.WhenCanTheLpaBeUsed = page.TaskInProgress
				} else {
					lpa.WhenCanTheLpaBeUsed = f.When
					lpa.Tasks.WhenCanTheLpaBeUsed = page.TaskCompleted
				}
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.Restrictions)
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
		AnswerLater: page.PostFormString(r, "answer-later") == "1",
		When:        page.PostFormString(r, "when"),
	}
}

func (f *whenCanTheLpaBeUsedForm) Validate() validation.List {
	var errors validation.List

	errors.String("when", "whenYourAttorneysCanUseYourLpa", f.When,
		validation.Select(page.UsedWhenRegistered, page.UsedWhenCapacityLost))

	return errors
}
