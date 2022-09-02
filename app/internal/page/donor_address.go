package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type donorAddressData struct {
	App       AppData
	Errors    map[string]string
	Addresses []Address
	Form      *donorAddressForm
}

func DonorAddress(logger Logger, tmpl template.Template, addressClient AddressClient, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &donorAddressData{
			App:  appData,
			Form: &donorAddressForm{},
		}

		if lpa.Donor.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.Donor.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readDonorAddressForm(r)
			data.Errors = data.Form.Validate()

			if (data.Form.Action == "manual" || data.Form.Action == "select") && len(data.Errors) == 0 {
				lpa.Donor.Address = *data.Form.Address
				dataStore.Put(r.Context(), appData.SessionID, lpa)
				appData.Lang.Redirect(w, r, howWouldYouLikeToBeContactedPath, http.StatusFound)
				return nil
			}

			if data.Form.Action == "lookup" && len(data.Errors) == 0 ||
				data.Form.Action == "select" && len(data.Errors) > 0 {
				addresses, err := addressClient.LookupPostcode(data.Form.LookupPostcode)
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
				data.Form.Address = &Address{}
			}
		}

		return tmpl(w, data)
	}
}

type donorAddressForm struct {
	Action         string
	LookupPostcode string
	Address        *Address
}

func readDonorAddressForm(r *http.Request) *donorAddressForm {
	d := &donorAddressForm{}
	d.Action = r.PostFormValue("action")

	switch d.Action {
	case "lookup":
		d.LookupPostcode = postFormString(r, "lookup-postcode")

	case "select":
		d.LookupPostcode = postFormString(r, "lookup-postcode")
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			d.Address = DecodeAddress(selectAddress)
		}

	case "manual":
		d.Address = &Address{
			Line1:      postFormString(r, "address-line-1"),
			Line2:      postFormString(r, "address-line-2"),
			TownOrCity: postFormString(r, "address-town"),
			Postcode:   postFormString(r, "address-postcode"),
		}
	}

	return d
}

func (d *donorAddressForm) Validate() map[string]string {
	errors := map[string]string{}

	switch d.Action {
	case "lookup":
		if d.LookupPostcode == "" {
			errors["lookup-postcode"] = "enterPostcode"
		}

	case "select":
		if d.Address == nil {
			errors["select-address"] = "selectAddress"
		}

	case "manual":
		if d.Address.Line1 == "" {
			errors["address-line-1"] = "enterAddress"
		}
		if d.Address.TownOrCity == "" {
			errors["address-town"] = "enterTownOrCity"
		}
		if d.Address.Postcode == "" {
			errors["address-postcode"] = "enterPostcode"
		}
	}

	return errors
}
