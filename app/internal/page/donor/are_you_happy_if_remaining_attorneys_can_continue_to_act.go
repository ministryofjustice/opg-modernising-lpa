package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfRemainingAttorneysCanContinueToActData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfRemainingAttorneysCanContinueToAct(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfRemainingAttorneysCanContinueToActData{
			App:   appData,
			Happy: lpa.HowAttorneysMakeDecisions.HappyIfRemainingCanContinueToAct,
		}

		if r.Method == http.MethodPost {
			form := readAreYouHappyIfOneAttorneyCantActNoneCanForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.HowAttorneysMakeDecisions.HappyIfRemainingCanContinueToAct = form.Happy
				lpa.Tasks.ChooseAttorneys = page.TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.DoYouWantReplacementAttorneys)
			}
		}

		return tmpl(w, data)
	}
}

type areYouHappyIfRemainingAttorneysCanContinueToActForm struct {
	Happy string
}

func readAreYouHappyIfRemainingAttorneysCanContinueToActForm(r *http.Request) *areYouHappyIfRemainingAttorneysCanContinueToActForm {
	return &areYouHappyIfRemainingAttorneysCanContinueToActForm{
		Happy: page.PostFormString(r, "happy"),
	}
}

func (f *areYouHappyIfRemainingAttorneysCanContinueToActForm) Validate() validation.List {
	var errors validation.List

	errors.String("happy", "yesIfYouAreHappyIfRemainingAttorneysCanContinueToAct", f.Happy,
		validation.Select("yes", "no"))

	return errors
}
