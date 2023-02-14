package appForm

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type AddressForm struct {
	Action         string
	LookupPostcode string
	Address        *place.Address
}

func ReadAddressForm(r *http.Request) *AddressForm {
	f := &AddressForm{}
	f.Action = r.PostFormValue("action")

	switch f.Action {
	case "lookup":
		f.LookupPostcode = page.PostFormString(r, "lookup-postcode")

	case "select":
		f.LookupPostcode = page.PostFormString(r, "lookup-postcode")
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			f.Address = page.DecodeAddress(selectAddress)
		}

	case "manual":
		f.Address = &place.Address{
			Line1:      page.PostFormString(r, "address-line-1"),
			Line2:      page.PostFormString(r, "address-line-2"),
			Line3:      page.PostFormString(r, "address-line-3"),
			TownOrCity: page.PostFormString(r, "address-town"),
			Postcode:   page.PostFormString(r, "address-postcode"),
		}
	}

	return f
}

func (f *AddressForm) Validate() validation.List {
	var errors validation.List

	switch f.Action {
	case "lookup":
		errors.String("lookup-postcode", "aPostcode", f.LookupPostcode,
			validation.Empty())

	case "select":
		errors.Address("select-address", "anAddressFromTheList", f.Address,
			validation.Selected())

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
		errors.String("address-postcode", "aPostcode", f.Address.Postcode,
			validation.Empty())
	}

	return errors
}
