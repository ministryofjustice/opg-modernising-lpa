package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whenCanTheLpaBeUsedData struct {
	App    page.AppData
	Errors validation.List
	When   string
	Lpa    *page.Lpa
}

func WhenCanTheLpaBeUsed(tmpl template.Template, donorStore DonorStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whenCanTheLpaBeUsedData{
			App:  appData,
			When: lpa.WhenCanTheLpaBeUsed,
			Lpa:  lpa,
		}

		if r.Method == http.MethodPost {
			form := readWhenCanTheLpaBeUsedForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.WhenCanTheLpaBeUsed = form.When
				lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}

type whenCanTheLpaBeUsedForm struct {
	When string
}

func readWhenCanTheLpaBeUsedForm(r *http.Request) *whenCanTheLpaBeUsedForm {
	return &whenCanTheLpaBeUsedForm{
		When: page.PostFormString(r, "when"),
	}
}

func (f *whenCanTheLpaBeUsedForm) Validate() validation.List {
	var errors validation.List

	errors.String("when", "whenYourAttorneysCanUseYourLpa", f.When,
		validation.Select(page.UsedWhenRegistered, page.UsedWhenCapacityLost))

	return errors
}
