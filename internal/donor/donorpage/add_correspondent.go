package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type addCorrespondentData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func AddCorrespondent(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &addCorrespondentData{
			App:   appData,
			Donor: donor,
			Form:  form.NewYesNoForm(donor.AddCorrespondent),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddCorrespondent")
			data.Errors = f.Validate()

			if data.Errors.None() {
				donor.AddCorrespondent = f.YesNo

				var redirectUrl page.LpaPath
				if donor.AddCorrespondent.IsNo() {
					donor.Correspondent = donordata.Correspondent{}
					donor.Tasks.AddCorrespondent = actor.TaskCompleted
					redirectUrl = page.Paths.TaskList
				} else {
					if donor.Correspondent.FirstNames == "" {
						donor.Tasks.AddCorrespondent = actor.TaskInProgress
					}
					redirectUrl = page.Paths.EnterCorrespondentDetails
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirectUrl.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
