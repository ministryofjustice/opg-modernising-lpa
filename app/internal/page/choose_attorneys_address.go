package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseAttorneysAddressData struct {
	App       AppData
	Errors    map[string]string
	Attorney  Attorney
	Addresses []place.Address
	Form      *chooseAttorneysAddressForm
}

func ChooseAttorneysAddress(logger Logger, tmpl template.Template, addressClient AddressClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &chooseAttorneysAddressData{
			App:      appData,
			Attorney: lpa.Attorney,
			Form:     &chooseAttorneysAddressForm{},
		}

		if lpa.Attorney.Address.Line1 != "" {
			data.Form.Action = "manual"
			data.Form.Address = &lpa.Attorney.Address
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysAddressForm(r)
			data.Errors = data.Form.Validate()

			if (data.Form.Action == "manual" || data.Form.Action == "select") && len(data.Errors) == 0 {
				lpa.Attorney.Address = *data.Form.Address
				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, wantReplacementAttorneysPath, http.StatusFound)
				return nil
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

type chooseAttorneysAddressForm struct {
	Action         string
	LookupPostcode string
	Address        *place.Address
}

func readChooseAttorneysAddressForm(r *http.Request) *chooseAttorneysAddressForm {
	d := &chooseAttorneysAddressForm{}
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
		d.Address = &place.Address{
			Line1:      postFormString(r, "address-line-1"),
			Line2:      postFormString(r, "address-line-2"),
			Line3:      postFormString(r, "address-line-3"),
			TownOrCity: postFormString(r, "address-town"),
			Postcode:   postFormString(r, "address-postcode"),
		}
	}

	return d
}

func (d *chooseAttorneysAddressForm) Validate() map[string]string {
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
	}

	return errors
}
