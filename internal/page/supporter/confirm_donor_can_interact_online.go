package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDonorCanInteractOnlineData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
}

func ConfirmDonorCanInteractOnline(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		data := &confirmDonorCanInteractOnlineData{
			App:  appData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "ifYouWouldLikeToContinueMakingAnOnlineLPA")
			data.Errors = data.Form.Validate()

			if data.Form.YesNo.IsYes() {
				donorProvided, err := organisationStore.CreateLPA(r.Context())
				if err != nil {
					return err
				}

				return page.Paths.YourName.Redirect(w, r, appData, donorProvided)
			} else if data.Form.YesNo.IsNo() {
				return page.Paths.Supporter.ContactOPGForPaperForms.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}
