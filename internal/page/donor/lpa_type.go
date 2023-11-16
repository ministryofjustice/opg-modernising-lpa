package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaTypeOptions struct {
	PropertyFinance page.LpaType
	HealthWelfare   page.LpaType
}

type lpaTypeData struct {
	App     page.AppData
	Errors  validation.List
	Form    *lpaTypeForm
	Options lpaTypeOptions
}

func LpaType(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &lpaTypeData{
			App: appData,
			Form: &lpaTypeForm{
				LpaType: lpa.Type,
			},
			Options: lpaTypeOptions{
				PropertyFinance: page.LpaTypePropertyFinance,
				HealthWelfare:   page.LpaTypeHealthWelfare,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readLpaTypeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if lpa.Type != data.Form.LpaType {
					lpa.Type = data.Form.LpaType
					lpa.Tasks.YourDetails = actor.TaskCompleted
					lpa.HasSentApplicationUpdatedEvent = false

					if err := donorStore.Put(r.Context(), lpa); err != nil {
						return err
					}
				}

				return page.Paths.TaskList.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type lpaTypeForm struct {
	LpaType page.LpaType
	Error   error
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	lpaType, err := page.ParseLpaType(page.PostFormString(r, "lpa-type"))

	return &lpaTypeForm{
		LpaType: lpaType,
		Error:   err,
	}
}

func (f *lpaTypeForm) Validate() validation.List {
	var errors validation.List

	errors.Error("lpa-type", "theTypeOfLpaToMake", f.Error,
		validation.Selected())

	return errors
}
