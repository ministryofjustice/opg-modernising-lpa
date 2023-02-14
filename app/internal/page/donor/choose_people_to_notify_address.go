package donor

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifyAddressData struct {
	App            page.AppData
	Errors         validation.List
	PersonToNotify actor.PersonToNotify
	Addresses      []place.Address
	Form           *form.AddressForm
}

func ChoosePeopleToNotifyAddress(logger page.Logger, tmpl template.Template, addressClient page.AddressClient, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		personId := r.FormValue("id")
		personToNotify, found := lpa.PeopleToNotify.Get(personId)

		if found == false {
			return appData.Redirect(w, r, lpa, page.Paths.ChoosePeopleToNotify)
		}

		data := &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form:           &form.AddressForm{},
		}

		if personToNotify.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &personToNotify.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.Action == "manual" && data.Errors.None() {
				personToNotify.Address = *data.Form.Address
				lpa.PeopleToNotify.Put(personToNotify)
				lpa.Tasks.PeopleToNotify = page.TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = appData.Paths.ChoosePeopleToNotifySummary
				}

				return appData.Redirect(w, r, lpa, from)
			}

			// Force the manual address view after selecting
			if data.Form.Action == "select" && data.Errors.None() {
				data.Form.Action = "manual"

				personToNotify.Address = *data.Form.Address
				lpa.PeopleToNotify.Put(personToNotify)

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}
			}

			if data.Form.Action == "lookup" && data.Errors.None() ||
				data.Form.Action == "select" && data.Errors.Any() {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)

					if errors.As(err, &place.BadRequestError{}) {
						data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
					} else {
						data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
					}
				} else if len(addresses) == 0 {
					data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noAddressesFound"})
				}

				data.Addresses = addresses
			}
		}

		if r.Method == http.MethodGet {
			action := r.FormValue("action")
			if action == "manual" {
				data.Form.Action = "manual"
				data.Form.Address = &place.Address{}
			}
		}

		return tmpl(w, data)
	}
}
