package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type areYouHappyIfOneReplacementAttorneyCantActNoneCanData struct {
	App    page.AppData
	Errors validation.List
	Happy  string
}

func AreYouHappyIfOneReplacementAttorneyCantActNoneCan(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &areYouHappyIfOneReplacementAttorneyCantActNoneCanData{
			App:   appData,
			Happy: lpa.HowReplacementAttorneysMakeDecisions.HappyIfOneCannotActNoneCan,
		}

		if r.Method == http.MethodPost {
			form := readAreYouHappyIfOneAttorneyCantActNoneCanForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.HowReplacementAttorneysMakeDecisions.HappyIfOneCannotActNoneCan = form.Happy

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if form.Happy == "yes" {
					return appData.Redirect(w, r, lpa, page.Paths.WhenCanTheLpaBeUsed)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.AreYouHappyIfRemainingReplacementAttorneysCanContinueToAct)
				}
			}
		}

		return tmpl(w, data)
	}
}

type areYouHappyIfOneReplacementAttorneyCantActNoneCanForm struct {
	Happy string
}

func readAreYouHappyIfOneReplacementAttorneyCantActNoneCanForm(r *http.Request) *areYouHappyIfOneReplacementAttorneyCantActNoneCanForm {
	return &areYouHappyIfOneReplacementAttorneyCantActNoneCanForm{
		Happy: page.PostFormString(r, "happy"),
	}
}

func (f *areYouHappyIfOneReplacementAttorneyCantActNoneCanForm) Validate() validation.List {
	var errors validation.List

	errors.String("happy", "yesIfYouAreHappyIfOneReplacementAttorneyCantActNoneCan", f.Happy,
		validation.Select("yes", "no"))

	return errors
}
