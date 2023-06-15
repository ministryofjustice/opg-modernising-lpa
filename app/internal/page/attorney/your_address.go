package attorney

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type yourAddressData struct {
	App       page.AppData
	Errors    validation.List
	Addresses []place.Address
	Form      *form.AddressForm
}

func YourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, attorneyStore AttorneyStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		data := &yourAddressData{
			App:  appData,
			Form: &form.AddressForm{},
		}

		if attorneyProvidedDetails.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &attorneyProvidedDetails.Address
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadAddressForm(r)
			data.Errors = data.Form.Validate(true)

			lookupAddress := func() {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)

					if errors.As(err, &place.BadRequestError{}) {
						data.Errors.Add("lookup-postcode", validation.EnterError{Label: "invalidPostcode"})
					} else {
						data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
					}

					data.Form.Action = "postcode"
				} else if len(addresses) == 0 {
					data.Errors.Add("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"})

					data.Form.Action = "postcode"
				}

				data.Addresses = addresses
			}

			switch data.Form.Action {
			case "manual":
				if data.Errors.None() {
					attorneyProvidedDetails.Address = *data.Form.Address
					attorneyProvidedDetails.Tasks.ConfirmYourDetails = actor.TaskCompleted

					if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
						return err
					}

					return appData.Redirect(w, r, nil, appData.Paths.Attorney.ReadTheLpa)
				}

			case "postcode-select":
				if data.Errors.None() {
					data.Form.Action = "manual"
				} else {
					lookupAddress()
				}

			case "postcode-lookup":
				if data.Errors.None() {
					lookupAddress()
				} else {
					data.Form.Action = "postcode"
				}
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
