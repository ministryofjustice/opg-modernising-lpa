package form

var FieldNames = SharedFieldNames{
	LanguagePreference: LanguagePreferenceFieldNames{
		LanguagePreference: "language-preference",
	},
	Address: AddressFieldNames{
		Line1:      "address-line-1",
		Line2:      "address-line-2",
		Line3:      "address-line-3",
		TownOrCity: "address-town",
		Postcode:   "address-postcode",
		Action:     "action",
	},
}

type SharedFieldNames struct {
	Address            AddressFieldNames
	LanguagePreference LanguagePreferenceFieldNames
}

type AddressFieldNames struct {
	Line1      string
	Line2      string
	Line3      string
	TownOrCity string
	Postcode   string
	Action     string
}

type LanguagePreferenceFieldNames struct {
	LanguagePreference string
}
