package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func EnterCorrespondentAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := newChooseAddressData(
			appData,
			"correspondent",
			provided.Correspondent.FullName(),
			actoruid.UID{},
		)

		if provided.Correspondent.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &provided.Correspondent.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				provided.Tasks.AddCorrespondent = task.StateCompleted
				provided.Correspondent.Address = address

				if err := eventClient.SendCorrespondentUpdated(r.Context(), event.CorrespondentUpdated{
					UID:        provided.LpaUID,
					ActorUID:   &provided.Correspondent.UID,
					FirstNames: provided.Correspondent.FirstNames,
					LastName:   provided.Correspondent.LastName,
					Email:      provided.Correspondent.Email,
					Phone:      provided.Correspondent.Phone,
					Address:    &provided.Correspondent.Address,
				}); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathTaskList.Redirect(w, r, appData, provided)
			}

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					return setAddress(*data.Form.Address)
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress(r.Context(), logger, addressClient, data, false)
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress(r.Context(), logger, addressClient, data, false)
				} else {
					data.Form.Action = "postcode"
				}

			case "reuse":
				data.Addresses = provided.ActorAddresses()

			case "reuse-select":
				if data.Errors.None() {
					return setAddress(*data.Form.Address)
				} else {
					data.Addresses = provided.ActorAddresses()
				}
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue(data.Form.FieldNames.Action)
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
