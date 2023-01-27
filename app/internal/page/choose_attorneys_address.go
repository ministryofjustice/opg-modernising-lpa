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

			if data.Form.Action == "manual" && data.Errors.Empty() {
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
			if data.Form.Action == "select" && data.Errors.Empty() {
				data.Form.Action = "manual"

				attorney.Address = *data.Form.Address
				lpa.PutAttorney(attorney)

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}
			}

			if data.Form.Action == "lookup" && data.Errors.Empty() ||
				data.Form.Action == "select" && data.Errors.Any() {
				addresses, err := addressClient.LookupPostcode(r.Context(), data.Form.LookupPostcode)
				if err != nil {
					logger.Print(err)
					data.Errors.Add("lookup-postcode", "couldNotLookupPostcode")
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

func (d *chooseAttorneysAddressForm) Validate() validation.List {
	var errors validation.List

	switch d.Action {
	case "lookup":
		if d.LookupPostcode == "" {
			errors.Add("lookup-postcode", "enterPostcode")
		}

	case "select":
		if d.Address == nil {
			errors.Add("select-address", "selectAddress")
		}

	case "manual":
		if d.Address.Line1 == "" {
			errors.Add("address-line-1", "enterAddress")
		}
		if len(d.Address.Line1) > 50 {
			errors.Add("address-line-1", "addressLine1TooLong")
		}
		if len(d.Address.Line2) > 50 {
			errors.Add("address-line-2", "addressLine2TooLong")
		}
		if len(d.Address.Line3) > 50 {
			errors.Add("address-line-3", "addressLine3TooLong")
		}
		if d.Address.TownOrCity == "" {
			errors.Add("address-town", "enterTownOrCity")
		}
	}

	return errors
}
