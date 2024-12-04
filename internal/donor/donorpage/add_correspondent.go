package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type addCorrespondentData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func AddCorrespondent(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &addCorrespondentData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(provided.AddCorrespondent),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddCorrespondent")
			data.Errors = f.Validate()

			if data.Errors.None() {
				provided.AddCorrespondent = f.YesNo

				var redirectUrl donor.Path
				if provided.AddCorrespondent.IsNo() {
					if provided.Correspondent.FirstNames != "" {
						if err := eventClient.SendCorrespondentUpdated(r.Context(), event.CorrespondentUpdated{
							UID: provided.LpaUID,
						}); err != nil {
							return err
						}
					}

					provided.Correspondent = donordata.Correspondent{}
					provided.Tasks.AddCorrespondent = task.StateCompleted

					redirectUrl = donor.PathTaskList
				} else {
					if provided.Correspondent.FirstNames == "" {
						provided.Tasks.AddCorrespondent = task.StateInProgress
					}
					redirectUrl = donor.PathEnterCorrespondentDetails
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return redirectUrl.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
