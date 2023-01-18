package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type yourAddressData struct {
	App       AppData
	Errors    map[string]string
	Addresses []place.Address
	Form      *yourAddressForm
}

func YourAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &yourAddressData{
			App:  appData,
			Form: &yourAddressForm{},
		}

		if lpa.You.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.You.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readYourAddressForm(r)
			data.Errors = data.Form.Validate()

			if data.Form.Action == "manual" && len(data.Errors) == 0 {
				lpa.You.Address = *data.Form.Address
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, lpa, Paths.WhoIsTheLpaFor)
			}

			if data.Form.Action == "select" && len(data.Errors) == 0 {
				data.Form.Action = "manual"
			}

			if data.Form.Action == "lookup" && len(data.Errors) == 0 ||
				data.Form.Action == "select" && len(data.Errors) > 0 {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)
					data.Errors["lookup-postcode"] = "couldNotLookupPostcode"
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

type yourAddressForm struct {
	Action         string
	LookupPostcode place.Postcode
	Address        *place.Address
}

func readYourAddressForm(r *http.Request) *yourAddressForm {
	d := &yourAddressForm{}
	d.Action = r.PostFormValue("action")

	switch d.Action {
	case "lookup":
		d.LookupPostcode = place.Postcode(postFormString(r, "lookup-postcode"))

	case "select":
		d.LookupPostcode = place.Postcode(postFormString(r, "lookup-postcode"))
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			d.Address = DecodeAddress(selectAddress)
		}

	case "manual":
		d.Address = &place.Address{
			Line1:      postFormString(r, "address-line-1"),
			Line2:      postFormString(r, "address-line-2"),
			Line3:      postFormString(r, "address-line-3"),
			TownOrCity: postFormString(r, "address-town"),
			Postcode:   place.Postcode(postFormString(r, "address-postcode")),
		}
	}

	return d
}

func (d *yourAddressForm) Validate() map[string]string {
	errors := map[string]string{}

	switch d.Action {
	case "lookup":
		if d.LookupPostcode == "" {
			errors["lookup-postcode"] = "enterPostcode"
		}
		if !d.LookupPostcode.IsUkFormat() {
			errors["lookup-postcode"] = "enterUkPostcode"
		}

	case "select":
		if d.Address == nil {
			errors["select-address"] = "selectAddress"
		}

	case "manual":
		if d.Address.Line1 == "" {
			errors["address-line-1"] = "enterAddress"
		}
		if len(d.Address.Line1) > 50 {
			errors["address-line-1"] = "addressLine1TooLong"
		}
		if len(d.Address.Line2) > 50 {
			errors["address-line-2"] = "addressLine2TooLong"
		}
		if len(d.Address.Line3) > 50 {
			errors["address-line-3"] = "addressLine3TooLong"
		}
		if d.Address.TownOrCity == "" {
			errors["address-town"] = "enterTownOrCity"
		}
		if d.Address.Postcode == "" {
			errors["address-postcode"] = "enterPostcode"
		}
		if !d.Address.Postcode.IsUkFormat() {
			errors["address-postcode"] = "enterUkPostcode"
		}
	}

	return errors
}
