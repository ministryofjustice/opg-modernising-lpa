package newforms

type AddressForm struct {
	Action *String

	// postcode-lookup
	LookupPostcode *String

	// postcode-select|reuse-select
	SelectAddress *String

	// manual
	Line1      *String
	Line2      *String
	Line3      *String
	TownOrCity *String
	Postcode   *String
}

func NewAddressForm(l Localizer, useYour bool) *AddressForm {
	postcodeLabel := l.T("aPostcode")
	selectAddressLabel := l.T("anAddressFromTheList")
	line1Label := l.T("addressLine1")
	line2Label := l.T("addressLine2")
	line3Label := l.T("addressLine3")
	townOrCityLabel := l.T("townOrCity")
	if useYour {
		postcodeLabel = l.T("yourPostcode")
		selectAddressLabel = l.T("yourAddressFromTheList")
		line1Label = l.T("addressLine1OfYourAddress")
		line2Label = l.T("addressLine2OfYourAddress")
		line3Label = l.T("addressLine3OfYourAddress")
		townOrCityLabel = l.T("yourTownOrCity")
	}

	return &AddressForm{
		Action:         NewString("action", ""),
		LookupPostcode: NewString("lookup-postcode", postcodeLabel),
		SelectAddress:  NewString("select-address", selectAddressLabel),
		Line1:          NewString("line-1", line1Label),
		Line2:          NewString("line-2", line2Label),
		Line3:          NewString("line-3", line3Label),
		TownOrCity:     NewString("town", townOrCityLabel),
	}
}
