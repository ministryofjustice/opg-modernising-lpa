package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type restrictionsData struct {
	App          AppData
	Errors       map[string]string
	Restrictions string
	Completed    bool
}

func Restrictions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &restrictionsData{
			App:          appData,
			Restrictions: lpa.Restrictions,
			Completed:    lpa.Tasks.Restrictions == TaskCompleted,
		}

		if r.Method == http.MethodPost {
			form := readRestrictionsForm(r)
			data.Errors = form.Validate()
			data.Restrictions = form.Restrictions

			if len(data.Errors) == 0 || form.AnswerLater {
				if form.AnswerLater {
					lpa.Tasks.Restrictions = TaskInProgress
				} else {
					lpa.Tasks.Restrictions = TaskCompleted
					lpa.Restrictions = form.Restrictions
				}
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance, http.StatusFound)
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

func (f *restrictionsForm) Validate() map[string]string {
	errors := map[string]string{}

	if len(f.Restrictions) > 10000 {
		errors["restrictions"] = "restrictionsTooLong"
	}

	return errors
}
