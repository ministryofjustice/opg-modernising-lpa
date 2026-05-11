package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

//go:generate go tool enumerator -type howYouWillConfirmYourIdentity -empty -trimprefix
type howYouWillConfirmYourIdentity uint8

const (
	howYouWillConfirmYourIdentityAtPostOffice howYouWillConfirmYourIdentity = iota + 1
	howYouWillConfirmYourIdentityPostOfficeSuccessfully
	howYouWillConfirmYourIdentityOneLogin
	howYouWillConfirmYourIdentityWithdraw
)

type howWillYouConfirmYourIdentityData struct {
	App  appcontext.Data
	Form *newforms.EnumForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
}

func HowWillYouConfirmYourIdentity(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howWillYouConfirmYourIdentityData{
			App:  appData,
			Form: newforms.NewEnumForm[howYouWillConfirmYourIdentity](appData.Localizer.T("howYouWillConfirmYourIdentity"), howYouWillConfirmYourIdentityValues),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			switch data.Form.Enum.Value {
			case howYouWillConfirmYourIdentityAtPostOffice:
				provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

				if err := eventClient.SendConfirmAtPostOfficeSelected(r.Context(), event.ConfirmAtPostOfficeSelected{
					UID: provided.LpaUID,
				}); err != nil {
					return fmt.Errorf("error sending event: %w", err)
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return fmt.Errorf("error updating donor: %w", err)
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)

			case howYouWillConfirmYourIdentityWithdraw:
				if provided.WitnessedByCertificateProviderAt.IsZero() {
					return donor.PathDeleteThisLpa.Redirect(w, r, appData, provided)
				}

				return donor.PathWithdrawThisLpa.Redirect(w, r, appData, provided)

			default:
				return donor.PathIdentityWithOneLogin.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
