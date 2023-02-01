package page

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
		Postcode:   "e",
	}

	testCases := map[string]struct {
		form   url.Values
		result *addressForm
	}{
		"lookup": {
			form: url.Values{
				"action":          {"lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &addressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"select": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &addressForm{
				Action:  "select",
				Address: expectedAddress,
			},
		},
		"select-not-selected": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {""},
			},
			result: &addressForm{
				Action:  "select",
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
			result: &addressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			actual := readAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *addressForm
		errors validation.List
	}{
		"lookup valid": {
			form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"lookup missing postcode": {
			form: &addressForm{
				Action: "lookup",
			},
			errors: validation.With("lookup-postcode", validation.EnterError{Label: "postcode"}),
		},
		"select valid": {
			form: &addressForm{
				Action:  "select",
				Address: &place.Address{},
			},
		},
		"select not selected": {
			form: &addressForm{
				Action:  "select",
				Address: nil,
			},
			errors: validation.With("select-address", validation.SelectError{Label: "address"}),
		},
		"manual valid": {
			form: &addressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      "a",
					TownOrCity: "b",
				},
			},
		},
		"manual missing all": {
			form: &addressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With("address-line-1", validation.EnterError{Label: "addressLine1"}).
				With("address-town", validation.EnterError{Label: "townOrCity"}),
		},
		"manual max length": {
			form: &addressForm{
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
			form: &addressForm{
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
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
