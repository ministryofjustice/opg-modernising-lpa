package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadAddressForm(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "E",
		Country:    "GB",
	}

	testCases := map[string]struct {
		form   url.Values
		result *AddressForm
	}{
		"postcode-lookup": {
			form: url.Values{
				"action":          {"postcode-lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
				FieldNames:     FieldNames.Address,
			},
		},
		"postcode-select": {
			form: url.Values{
				"action":         {"postcode-select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &AddressForm{
				Action:     "postcode-select",
				Address:    expectedAddress,
				FieldNames: FieldNames.Address,
			},
		},
		"postcode-select not selected": {
			form: url.Values{
				"action":         {"postcode-select"},
				"select-address": {""},
			},
			result: &AddressForm{
				Action:     "postcode-select",
				Address:    nil,
				FieldNames: FieldNames.Address,
			},
		},
		"reuse-select": {
			form: url.Values{
				"action":         {"reuse-select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &AddressForm{
				Action:     "reuse-select",
				Address:    expectedAddress,
				FieldNames: FieldNames.Address,
			},
		},
		"reuse-select not selected": {
			form: url.Values{
				"action":         {"reuse-select"},
				"select-address": {""},
			},
			result: &AddressForm{
				Action:     "reuse-select",
				Address:    nil,
				FieldNames: FieldNames.Address,
			},
		},
		"manual": {
			form: url.Values{
				"action":                      {"manual"},
				FieldNames.Address.Line1:      {"a"},
				FieldNames.Address.Line2:      {"b"},
				FieldNames.Address.Line3:      {"c"},
				FieldNames.Address.TownOrCity: {"d"},
				FieldNames.Address.Postcode:   {"e"},
			},
			result: &AddressForm{
				Action:     "manual",
				Address:    expectedAddress,
				FieldNames: FieldNames.Address,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			actual := ReadAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form    *AddressForm
		errors  validation.List
		useYour bool
	}{
		"action missing": {
			form:   &AddressForm{},
			errors: validation.With("action", validation.SelectError{Label: "ifUsePreviousAddressOrEnterNew"}),
		},
		"postcode-lookup valid": {
			form: &AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
			},
		},
		"postcode-lookup missing postcode": {
			form: &AddressForm{
				Action: "postcode-lookup",
			},
			errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		},
		"postcode-lookup your missing postcode": {
			form: &AddressForm{
				Action: "postcode-lookup",
			},
			errors:  validation.With("lookup-postcode", validation.EnterError{Label: "yourPostcode"}),
			useYour: true,
		},
		"postcode-select valid": {
			form: &AddressForm{
				Action:  "postcode-select",
				Address: &place.Address{},
			},
		},
		"postcode-select not selected": {
			form: &AddressForm{
				Action:  "postcode-select",
				Address: nil,
			},
			errors: validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		},
		"postcode-select your address not selected": {
			form: &AddressForm{
				Action:  "postcode-select",
				Address: nil,
			},
			errors:  validation.With("select-address", validation.SelectError{Label: "yourAddressFromTheList"}),
			useYour: true,
		},
		"reuse-select valid": {
			form: &AddressForm{
				Action:  "reuse-select",
				Address: &place.Address{},
			},
		},
		"reuse-select not selected": {
			form: &AddressForm{
				Action:  "reuse-select",
				Address: nil,
			},
			errors: validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		},
		"reuse-select your address not selected": {
			form: &AddressForm{
				Action:  "reuse-select",
				Address: nil,
			},
			errors:  validation.With("select-address", validation.SelectError{Label: "yourAddressFromTheList"}),
			useYour: true,
		},
		"manual valid": {
			form: &AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      "a",
					TownOrCity: "b",
					Postcode:   "C12 1CC",
				},
			},
		},
		"manual missing all": {
			form: &AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With(FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1"}).
				With(FieldNames.Address.TownOrCity, validation.EnterError{Label: "townOrCity"}).
				With(FieldNames.Address.Postcode, validation.EnterError{Label: "aPostcode"}),
		},
		"manual missing all your": {
			form: &AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With(FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1OfYourAddress"}).
				With(FieldNames.Address.TownOrCity, validation.EnterError{Label: "yourTownOrCity"}).
				With(FieldNames.Address.Postcode, validation.EnterError{Label: "yourPostcode"}),
			useYour: true,
		},
		"manual max length": {
			form: &AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 50),
					Line2:      strings.Repeat("x", 50),
					Line3:      strings.Repeat("x", 50),
					TownOrCity: "b",
					Postcode:   "C",
				},
			},
		},
		"manual too long": {
			form: &AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 51),
					Line2:      strings.Repeat("x", 51),
					Line3:      strings.Repeat("x", 51),
					TownOrCity: "b",
					Postcode:   "C",
				},
			},
			errors: validation.
				With(FieldNames.Address.Line1, validation.StringTooLongError{Label: "addressLine1", Length: 50}).
				With(FieldNames.Address.Line2, validation.StringTooLongError{Label: "addressLine2Label", Length: 50}).
				With(FieldNames.Address.Line3, validation.StringTooLongError{Label: "addressLine3Label", Length: 50}),
		},
		"manual your too long": {
			form: &AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 51),
					Line2:      strings.Repeat("x", 51),
					Line3:      strings.Repeat("x", 51),
					TownOrCity: "b",
					Postcode:   "C",
				},
			},
			errors: validation.
				With(FieldNames.Address.Line1, validation.StringTooLongError{Label: "addressLine1OfYourAddress", Length: 50}).
				With(FieldNames.Address.Line2, validation.StringTooLongError{Label: "addressLine2OfYourAddress", Length: 50}).
				With(FieldNames.Address.Line3, validation.StringTooLongError{Label: "addressLine3OfYourAddress", Length: 50}),
			useYour: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.useYour))
		})
	}
}
