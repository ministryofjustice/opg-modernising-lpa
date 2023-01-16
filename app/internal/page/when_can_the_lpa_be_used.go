package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type whenCanTheLpaBeUsedData struct {
	App       AppData
	Errors    map[string]string
	When      string
	Completed bool
	Lpa       *Lpa
}

func WhenCanTheLpaBeUsed(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
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

			if len(data.Errors) == 0 || form.AnswerLater {
				if form.AnswerLater {
					lpa.Tasks.WhenCanTheLpaBeUsed = TaskInProgress
				} else {
					lpa.WhenCanTheLpaBeUsed = form.When
				}
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.Restrictions)
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

func (f *whenCanTheLpaBeUsedForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.When != UsedWhenRegistered && f.When != UsedWhenCapacityLost {
		errors["when"] = "selectWhenCanTheLpaBeUsed"
	}

	return errors
}
