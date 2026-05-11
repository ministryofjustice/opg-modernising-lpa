package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDonorCanInteractOnlineData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *newforms.YesNoForm
}

func ConfirmDonorCanInteractOnline(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		data := &confirmDonorCanInteractOnlineData{
			App:  appData,
			Form: newforms.NewYesNoForm("ifYouWouldLikeToContinueMakingAnOnlineLPA"),
		}

		if r.Method == http.MethodPost {
			if ok := data.Form.Parse(r); ok && data.Form.YesNo.Value.IsYes() {
				donorProvided, err := organisationStore.CreateLPA(r.Context())
				if err != nil {
					return err
				}

				return donor.PathYourName.Redirect(w, r, appData, donorProvided)
			} else if data.Form.YesNo.Value.IsNo() {
				return supporter.PathContactOPGForPaperForms.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}
