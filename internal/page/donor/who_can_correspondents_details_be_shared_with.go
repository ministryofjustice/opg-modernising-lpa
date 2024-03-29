package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoCanCorrespondentsDetailsBeSharedWithData struct {
	App     page.AppData
	Errors  validation.List
	Form    *whoCanCorrespondentsDetailsBeSharedWithForm
	Options actor.CorrespondentShareOptions
}

func WhoCanCorrespondentsDetailsBeSharedWith(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &whoCanCorrespondentsDetailsBeSharedWithData{
			App: appData,
			Form: &whoCanCorrespondentsDetailsBeSharedWithForm{
				Share: donor.Correspondent.Share,
			},
			Options: actor.CorrespondentShareValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhoCanCorrespondentsDetailsBeSharedWithForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Correspondent.Share = data.Form.Share
				donor.Tasks.AddCorrespondent = actor.TaskCompleted

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.TaskList.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type whoCanCorrespondentsDetailsBeSharedWithForm struct {
	Share actor.CorrespondentShare
	Error error
}

func readWhoCanCorrespondentsDetailsBeSharedWithForm(r *http.Request) *whoCanCorrespondentsDetailsBeSharedWithForm {
	r.ParseForm()
	share, err := actor.ParseCorrespondentShare(r.PostForm["share"])

	return &whoCanCorrespondentsDetailsBeSharedWithForm{
		Share: share,
		Error: err,
	}
}

func (f *whoCanCorrespondentsDetailsBeSharedWithForm) Validate() validation.List {
	var errors validation.List

	errors.Error("share", "whoCorrespondentDetailsCanBeSharedWith", f.Error,
		validation.Selected())

	return errors
}
