package page

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysAddressData struct {
	App       AppData
	Errors    validation.List
	Attorney  Attorney
	Addresses []place.Address
	Form      *addressForm
}

func ChooseAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneyId := r.FormValue("id")
		attorney, found := lpa.GetAttorney(attorneyId)

		if found == false {
			return appData.Redirect(w, r, lpa, Paths.ChooseAttorneys)
		}

		data := &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form:     &addressForm{},
		}

		if attorney.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &attorney.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.Action == "manual" && data.Errors.None() {
				attorney.Address = *data.Form.Address
				lpa.PutAttorney(attorney)
				lpa.Tasks.ChooseAttorneys = TaskCompleted

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")
				if from == "" {
					from = appData.Paths.ChooseAttorneysSummary
				}

				return appData.Redirect(w, r, lpa, from)
			}

			// Force the manual address view after selecting
			if data.Form.Action == "select" && data.Errors.None() {
				data.Form.Action = "manual"

				attorney.Address = *data.Form.Address
				lpa.PutAttorney(attorney)

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
