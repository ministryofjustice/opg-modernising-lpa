package page

import (
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
	Form      *chooseAttorneysAddressForm
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
			Form:     &chooseAttorneysAddressForm{},
		}

		if attorney.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &attorney.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysAddressForm(r)
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
					data.Errors.Add("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"})
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

type chooseAttorneysAddressForm struct {
	Action         string
	LookupPostcode string
	Address        *place.Address
}

func readChooseAttorneysAddressForm(r *http.Request) *chooseAttorneysAddressForm {
	f := &chooseAttorneysAddressForm{}
	f.Action = r.PostFormValue("action")

	switch f.Action {
	case "lookup":
		f.LookupPostcode = postFormString(r, "lookup-postcode")

	case "select":
		f.LookupPostcode = postFormString(r, "lookup-postcode")
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			f.Address = DecodeAddress(selectAddress)
		}

	case "manual":
		f.Address = &place.Address{
			Line1:      postFormString(r, "address-line-1"),
			Line2:      postFormString(r, "address-line-2"),
			Line3:      postFormString(r, "address-line-3"),
			TownOrCity: postFormString(r, "address-town"),
			Postcode:   postFormString(r, "address-postcode"),
		}
	}

	return f
}

func (f *chooseAttorneysAddressForm) Validate() validation.List {
	var errors validation.List

	switch f.Action {
	case "lookup":
		errors.String("lookup-postcode", "postcode", f.LookupPostcode,
			validation.Empty())

	case "select":
		errors.Address("select-address", "address", f.Address,
			validation.AddressSelected())

	case "manual":
		errors.String("address-line-1", "addressLine1", f.Address.Line1,
			validation.Empty(),
			validation.StringTooLong(50))
		errors.String("address-line-2", "addressLine2Label", f.Address.Line2,
			validation.StringTooLong(50))
		errors.String("address-line-3", "addressLine3Label", f.Address.Line3,
			validation.StringTooLong(50))
		errors.String("address-town", "townOrCity", f.Address.TownOrCity,
			validation.Empty())
	}

	return errors
}
