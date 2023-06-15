package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestReadAddressForm(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
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
			},
		},
		"postcode-select": {
			form: url.Values{
				"action":         {"postcode-select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &AddressForm{
				Action:  "postcode-select",
				Address: expectedAddress,
			},
		},
		"postcode-select not selected": {
			form: url.Values{
				"action":         {"postcode-select"},
				"select-address": {""},
			},
			result: &AddressForm{
				Action:  "postcode-select",
				Address: nil,
			},
		},
		"reuse-select": {
			form: url.Values{
				"action":         {"reuse-select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &AddressForm{
				Action:  "reuse-select",
				Address: expectedAddress,
			},
		},
		"reuse-select not selected": {
			form: url.Values{
				"action":         {"reuse-select"},
				"select-address": {""},
			},
			result: &AddressForm{
				Action:  "reuse-select",
				Address: nil,
			},
		},
		"manual": {
			form: url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-line-2":   {"b"},
				"address-line-3":   {"c"},
				"address-town":     {"d"},
				"address-postcode": {"e"},
			},
			result: &AddressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

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
			errors: validation.With("action", validation.SelectError{Label: "placeholder"}),
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
					Postcode:   "c",
				},
			},
		},
		"manual missing all": {
			form: &AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With("address-line-1", validation.EnterError{Label: "addressLine1"}).
				With("address-town", validation.EnterError{Label: "townOrCity"}).
				With("address-postcode", validation.EnterError{Label: "aPostcode"}),
		},
		"manual missing all your": {
			form: &AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With("address-line-1", validation.EnterError{Label: "addressLine1OfYourAddress"}).
				With("address-town", validation.EnterError{Label: "yourTownOrCity"}).
				With("address-postcode", validation.EnterError{Label: "yourPostcode"}),
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
					Postcode:   "c",
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
					Postcode:   "c",
				},
			},
			errors: validation.
				With("address-line-1", validation.StringTooLongError{Label: "addressLine1", Length: 50}).
				With("address-line-2", validation.StringTooLongError{Label: "addressLine2Label", Length: 50}).
				With("address-line-3", validation.StringTooLongError{Label: "addressLine3Label", Length: 50}),
		},
		"manual your too long": {
			form: &AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 51),
					Line2:      strings.Repeat("x", 51),
					Line3:      strings.Repeat("x", 51),
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
			errors: validation.
				With("address-line-1", validation.StringTooLongError{Label: "addressLine1OfYourAddress", Length: 50}).
				With("address-line-2", validation.StringTooLongError{Label: "addressLine2OfYourAddress", Length: 50}).
				With("address-line-3", validation.StringTooLongError{Label: "addressLine3OfYourAddress", Length: 50}),
			useYour: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.useYour))
		})
	}
}
