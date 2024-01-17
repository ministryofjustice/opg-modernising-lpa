package form

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type AddressForm struct {
	Action         string
	LookupPostcode string
	Address        *place.Address
	FieldNames     AddressFieldNames
}

func NewAddressForm() *AddressForm {
	return &AddressForm{
		FieldNames: FieldNames.Address,
	}
}

func ReadAddressForm(r *http.Request) *AddressForm {
	f := NewAddressForm()
	f.Action = r.PostFormValue("action")

	switch f.Action {
	case "postcode-lookup":
		f.LookupPostcode = PostFormString(r, "lookup-postcode")

	case "postcode-select":
		f.LookupPostcode = PostFormString(r, "lookup-postcode")
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			f.Address = DecodeAddress(selectAddress)
		}

	case "reuse-select":
		selectAddress := r.PostFormValue("select-address")
		if selectAddress != "" {
			f.Address = DecodeAddress(selectAddress)
		}

	case "manual":
		f.Address = &place.Address{
			Line1:      PostFormString(r, FieldNames.Address.Line1),
			Line2:      PostFormString(r, FieldNames.Address.Line2),
			Line3:      PostFormString(r, FieldNames.Address.Line3),
			TownOrCity: PostFormString(r, FieldNames.Address.TownOrCity),
			Postcode:   strings.ToUpper(PostFormString(r, FieldNames.Address.Postcode)),
			Country:    "GB",
		}
	}

	return f
}

func (f *AddressForm) Validate(useYour bool) validation.List {
	var errors validation.List

	errors.String("action", "ifUsePreviousAddressOrEnterNew", f.Action,
		validation.Select("reuse", "reuse-select", "postcode", "postcode-lookup", "postcode-select", "manual", "skip"))

	switch f.Action {
	case "postcode-lookup":
		if useYour {
			errors.String("lookup-postcode", "yourPostcode", f.LookupPostcode,
				validation.Empty())
		} else {
			errors.String("lookup-postcode", "aPostcode", f.LookupPostcode,
				validation.Empty())
		}

	case "postcode-select", "reuse-select":
		if useYour {
			errors.Address("select-address", "yourAddressFromTheList", f.Address,
				validation.Selected())
		} else {
			errors.Address("select-address", "anAddressFromTheList", f.Address,
				validation.Selected())
		}

	case "manual":
		if useYour {
			errors.String(FieldNames.Address.Line1, "addressLine1OfYourAddress", f.Address.Line1,
				validation.Empty(),
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.Line2, "addressLine2OfYourAddress", f.Address.Line2,
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.Line3, "addressLine3OfYourAddress", f.Address.Line3,
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.TownOrCity, "yourTownOrCity", f.Address.TownOrCity,
				validation.Empty())
			errors.String(FieldNames.Address.Postcode, "yourPostcode", f.Address.Postcode,
				validation.Empty(),
				validation.Postcode())
		} else {
			errors.String(FieldNames.Address.Line1, "addressLine1", f.Address.Line1,
				validation.Empty(),
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.Line2, "addressLine2Label", f.Address.Line2,
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.Line3, "addressLine3Label", f.Address.Line3,
				validation.StringTooLong(50))
			errors.String(FieldNames.Address.TownOrCity, "townOrCity", f.Address.TownOrCity,
				validation.Empty())
			errors.String(FieldNames.Address.Postcode, "aPostcode", f.Address.Postcode,
				validation.Empty(),
				validation.Postcode())
		}
	}

	return errors
}

func DecodeAddress(s string) *place.Address {
	var v place.Address
	json.Unmarshal([]byte(s), &v)
	return &v
}
