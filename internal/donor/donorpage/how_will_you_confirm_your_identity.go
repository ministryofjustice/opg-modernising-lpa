package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate enumerator -type howYouWillConfirmYourIdentity -empty -trimprefix
type howYouWillConfirmYourIdentity uint8

const (
	howYouWillConfirmYourIdentityAtPostOffice howYouWillConfirmYourIdentity = iota + 1
	howYouWillConfirmYourIdentityPostOfficeSuccessfully
	howYouWillConfirmYourIdentityOneLogin
	howYouWillConfirmYourIdentityWithdraw
)

type howWillYouConfirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.SelectForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
}

func HowWillYouConfirmYourIdentity(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howWillYouConfirmYourIdentityData{
			App:  appData,
			Form: form.NewEmptySelectForm[howYouWillConfirmYourIdentity](howYouWillConfirmYourIdentityValues, "howYouWillConfirmYourIdentity"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.Selected {
				case howYouWillConfirmYourIdentityAtPostOffice:
					provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

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
		}

		return tmpl(w, data)
	}
}
