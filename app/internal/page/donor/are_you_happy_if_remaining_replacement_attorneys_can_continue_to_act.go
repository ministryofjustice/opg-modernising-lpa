package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfRemainingReplacementAttorneysCanContinueToActData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfRemainingReplacementAttorneysCanContinueToActData{
			App:   appData,
			Happy: lpa.HowReplacementAttorneysMakeDecisions.HappyIfRemainingCanContinueToAct,
		}

		if r.Method == http.MethodPost {
			form := readAreYouHappyIfOneAttorneyCantActNoneCanForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.HowReplacementAttorneysMakeDecisions.HappyIfRemainingCanContinueToAct = form.Happy

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.WhenCanTheLpaBeUsed)
			}
		}

		return tmpl(w, data)
	}
}

type areYouHappyIfRemainingReplacementAttorneysCanContinueToActForm struct {
	Happy string
}

func readAreYouHappyIfRemainingReplacementAttorneysCanContinueToActForm(r *http.Request) *areYouHappyIfRemainingReplacementAttorneysCanContinueToActForm {
	return &areYouHappyIfRemainingReplacementAttorneysCanContinueToActForm{
		Happy: page.PostFormString(r, "happy"),
	}
}

func (f *areYouHappyIfRemainingReplacementAttorneysCanContinueToActForm) Validate() validation.List {
	var errors validation.List

	errors.String("happy", "yesIfYouAreHappyIfRemainingReplacementAttorneysCanContinueToAct", f.Happy,
		validation.Select("yes", "no"))

	return errors
}
