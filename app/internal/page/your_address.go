package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type yourAddressData struct {
	App       AppData
	Errors    map[string]string
	Addresses []Address
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

			if (data.Form.Action == "manual" || data.Form.Action == "select") && len(data.Errors) == 0 {
				lpa.You.Address = *data.Form.Address
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, whoIsTheLpaForPath, http.StatusFound)
				return nil
			}

			if data.Form.Action == "lookup" && len(data.Errors) == 0 ||
				data.Form.Action == "select" && len(data.Errors) > 0 {
				response, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)
					data.Errors["lookup-postcode"] = "couldNotLookupPostcode"
				}

				if response.TotalResults > 0 {
					data.Addresses = TransformAddressDetailsToAddresses(response.Results)
				} else {
					data.Addresses = []Address{}
				}
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

type yourAddressForm struct {
	Action         string
	LookupPostcode string
	Address        *Address
}

func readYourAddressForm(r *http.Request) *yourAddressForm {
	d := &yourAddressForm{}
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

func (d *yourAddressForm) Validate() map[string]string {
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
