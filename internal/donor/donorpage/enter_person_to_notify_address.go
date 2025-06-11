package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func EnterPersonToNotifyAddress(logger Logger, tmpl template.Template, addressClient AddressClient, donorStore DonorStore, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		personToNotify, found := provided.PeopleToNotify.Get(actoruid.FromRequest(r))

		if found == false {
			return donor.PathEnterPersonToNotify.Redirect(w, r, appData, provided)
		}

		data := newChooseAddressData(
			appData,
			"personToNotify",
			personToNotify.FullName(),
			personToNotify.UID,
		)

		if personToNotify.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &personToNotify.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(false)

			setAddress := func(address place.Address) error {
				personToNotify.Address = address
				provided.PeopleToNotify.Put(personToNotify)
				provided.Tasks.PeopleToNotify = task.StateCompleted

				if err := reuseStore.PutPersonToNotify(r.Context(), personToNotify); err != nil {
					return err
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
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
