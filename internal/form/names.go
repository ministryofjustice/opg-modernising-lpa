package form

var FieldNames = SharedFieldNames{
	LanguagePreference: "language-preference",
	Address: AddressFieldNames{
		Line1:      "address-line-1",
		Line2:      "address-line-2",
		Line3:      "address-line-3",
		TownOrCity: "address-town",
		Postcode:   "address-postcode",
		Action:     "action",
	},
	Select: "selected",
	YesNo:  "yes-no",
}

type SharedFieldNames struct {
	Address            AddressFieldNames
	LanguagePreference string
	Select             string
	YesNo              string
}

type AddressFieldNames struct {
	Line1      string
	Line2      string
	Line3      string
	TownOrCity string
	Postcode   string
	Action     string
}
