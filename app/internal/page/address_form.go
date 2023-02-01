package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type addressForm struct {
	Action         string
	LookupPostcode string
	Address        *place.Address
}

func readAddressForm(r *http.Request) *addressForm {
	f := &addressForm{}
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

func (f *addressForm) Validate() validation.List {
	var errors validation.List

	switch f.Action {
	case "lookup":
		errors.String("lookup-postcode", "postcode", f.LookupPostcode,
			validation.Empty())

	case "select":
		errors.Address("select-address", "address", f.Address,
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
	}

	return errors
}
