package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaTypeData struct {
	App    page.AppData
	Errors validation.List
	Type   string
}

func LpaType(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &lpaTypeData{
			App:  appData,
			Type: lpa.Type,
		}

		if r.Method == http.MethodPost {
			f := readLpaTypeForm(r)
			data.Errors = f.Validate()

			if data.Errors.None() {
				lpa.Tasks.YourDetails = page.TaskCompleted
				lpa.Type = f.LpaType
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList)
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
		LpaType: page.PostFormString(r, "lpa-type"),
	}
}

func (f *lpaTypeForm) Validate() validation.List {
	var errors validation.List

	errors.String("lpa-type", "theTypeOfLpaToMake", f.LpaType,
		validation.Select(page.LpaTypePropertyFinance, page.LpaTypeHealthWelfare, page.LpaTypeCombined))

	return errors
}
