package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaTypeData struct {
	App     page.AppData
	Errors  validation.List
	Form    *lpaTypeForm
	Options actor.LpaTypeOptions
}

func LpaType(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &lpaTypeData{
			App: appData,
			Form: &lpaTypeForm{
				LpaType: donor.Type,
			},
			Options: actor.LpaTypeValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readLpaTypeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if donor.Type != data.Form.LpaType {
					donor.Type = data.Form.LpaType
					if donor.Type.IsPersonalWelfare() {
						donor.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenCapacityLost
					} else {
						donor.WhenCanTheLpaBeUsed = actor.CanBeUsedWhenUnknown
					}
					donor.Tasks.YourDetails = actor.TaskCompleted
					donor.HasSentApplicationUpdatedEvent = false

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type lpaTypeForm struct {
	LpaType actor.LpaType
	Error   error
}

func readLpaTypeForm(r *http.Request) *lpaTypeForm {
	lpaType, err := actor.ParseLpaType(page.PostFormString(r, "lpa-type"))

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
