package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaTypeData struct {
	App    AppData
	Errors validation.List
	Type   string
}

func LpaType(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &lpaTypeData{
			App:  appData,
			Type: lpa.Type,
		}

		if r.Method == http.MethodPost {
			form := readLpaTypeForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.Tasks.YourDetails = TaskCompleted
				lpa.Type = form.LpaType
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}

type lpaTypeForm struct {
	LpaType string
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	return &lpaTypeForm{
		LpaType: postFormString(r, "lpa-type"),
	}
}

func (f *lpaTypeForm) Validate() validation.List {
	var errors validation.List

	if f.LpaType != LpaTypePropertyFinance && f.LpaType != LpaTypeHealthWelfare && f.LpaType != LpaTypeCombined {
		errors.Add("lpa-type", "selectLpaType")
	}

	return errors
}
