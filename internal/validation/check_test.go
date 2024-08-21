package validation

import (
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestCheckError(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    error
		checks   []ErrorChecker
		expected List
	}{
		"selected": {
			input:  nil,
			checks: []ErrorChecker{Selected()},
		},
		"selected invalid": {
			input:    errors.New("err"),
			checks:   []ErrorChecker{Selected()},
			expected: With(name, SelectError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.Error(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}

func TestCheckString(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    string
		checks   []StringChecker
		expected List
	}{
		"empty": {
			input:  "hello there",
			checks: []StringChecker{Empty()},
		},
		"empty invalid": {
			input:    "",
			checks:   []StringChecker{Empty()},
			expected: With(name, EnterError{Label: label}),
		},
		"select": {
			input:  "one",
			checks: []StringChecker{Select("one", "other")},
		},
		"select invalid": {
			input:    "none",
			checks:   []StringChecker{Select("one", "other")},
			expected: With(name, SelectError{Label: label}),
		},
		"select invalid custom": {
			input:    "none",
			checks:   []StringChecker{Select("one", "other").CustomError()},
			expected: With(name, CustomError{Label: label}),
		},
		"string too long": {
			input:  "hello ther",
			checks: []StringChecker{Empty(), StringTooLong(10)},
		},
		"string too long invalid": {
			input:    "hello there",
			checks:   []StringChecker{Empty(), StringTooLong(10)},
			expected: With(name, StringTooLongError{Label: label, Length: 10}),
		},
		"string length": {
			input:  "hello ther",
			checks: []StringChecker{Empty(), StringLength(10)},
		},
		"string length invalid": {
			input:    "hello there",
			checks:   []StringChecker{Empty(), StringLength(10)},
			expected: With(name, StringLengthError{Label: label, Length: 10}),
		},
		"mobile": {
			input:  "07777777777",
			checks: []StringChecker{Mobile()},
		},
		"mobile with spaces": {
			input:  " 0 7 7 7 7 7 7 7 7 7 7 ",
			checks: []StringChecker{Mobile()},
		},
		"mobile uk prefix": {
			input:  "+447777777777",
			checks: []StringChecker{Mobile()},
		},
		"mobile uk prefix with spaces": {
			input:  " + 4 4 7 7 7 7 7 7 7 7 7 7 ",
			checks: []StringChecker{Mobile()},
		},
		"mobile invalid too long": {
			input:    "01152222222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid too short": {
			input:    "011522222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid alpha chars": {
			input:    "01152a2222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid uk prefix too long": {
			input:    "+441152222222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid uk prefix too short": {
			input:    "+4411522222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid uk prefix alpha chars": {
			input:    "+441152a2222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"mobile invalid error label": {
			input:    "01152222222",
			checks:   []StringChecker{Mobile().ErrorLabel("this")},
			expected: With(name, CustomError{Label: "this"}),
		},
		"non uk mobile": {
			input:  "+337777777777",
			checks: []StringChecker{NonUKMobile()},
		},
		"non uk mobile with spaces": {
			input:  " + 3 3 7 7 7 7 7 7 7 7 7 7 ",
			checks: []StringChecker{NonUKMobile()},
		},
		"non uk mobile no prefix": {
			input:    "337777777777",
			checks:   []StringChecker{NonUKMobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"non uk mobile too long": {
			input:    "+3377777777777777",
			checks:   []StringChecker{NonUKMobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"non uk mobile too short": {
			input:    "+337",
			checks:   []StringChecker{NonUKMobile()},
			expected: With(name, PhoneError{Tmpl: "errorMobile", Label: label}),
		},
		"phone": {
			input:  "+337777777777",
			checks: []StringChecker{Phone()},
		},
		"phone with spaces": {
			input:  " + 3 3 7 7 7 7 7 7 7 7 7 7 ",
			checks: []StringChecker{Phone()},
		},
		"phone no prefix": {
			input:  "337777777777",
			checks: []StringChecker{Phone()},
		},
		"phone too long": {
			input:    "+3377777777777777",
			checks:   []StringChecker{Phone()},
			expected: With(name, PhoneError{Tmpl: "errorPhone", Label: label}),
		},
		"phone too short": {
			input:    "+337",
			checks:   []StringChecker{Phone()},
			expected: With(name, PhoneError{Tmpl: "errorPhone", Label: label}),
		},
		"postcode": {
			input:  "B14 7ET",
			checks: []StringChecker{Postcode()},
		},
		"postcode no spaces": {
			input:  "B147ET",
			checks: []StringChecker{Postcode()},
		},
		"postcode too long": {
			input:    "B12345678T",
			checks:   []StringChecker{Postcode()},
			expected: With(name, PostcodeError{Label: label}),
		},
		"postcode lowercase": {
			input:    "B14 7Et",
			checks:   []StringChecker{Postcode()},
			expected: With(name, PostcodeError{Label: label}),
		},
		"email": {
			input:  "name@example.com",
			checks: []StringChecker{Email()},
		},
		"empty email is valid": {
			input:  "",
			checks: []StringChecker{Email()},
		},
		"email invalid": {
			input:    "example.com",
			checks:   []StringChecker{Email()},
			expected: With(name, EmailError{Label: label}),
		},
		"reference number modernised": {
			input:  "M",
			checks: []StringChecker{ReferenceNumber()},
		},
		"reference number old": {
			input:  "7",
			checks: []StringChecker{ReferenceNumber()},
		},
		"reference number invalid": {
			input:    "a",
			checks:   []StringChecker{ReferenceNumber()},
			expected: With(name, ReferenceNumberError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.String(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}

func TestCheckDate(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    date.Date
		checks   []DateChecker
		expected List
	}{
		"date missing": {
			input:  date.New("2006", "1", "2"),
			checks: []DateChecker{DateMissing()},
		},
		"date missing invalid": {
			input:    date.New("", "", "10"),
			checks:   []DateChecker{DateMissing()},
			expected: With(name, DateMissingError{Label: label, MissingMonth: true, MissingYear: true}),
		},
		"date missing invalid all": {
			input:    date.New("", "", ""),
			checks:   []DateChecker{DateMissing()},
			expected: With(name, EnterError{Label: label}),
		},
		"date must be real": {
			input:  date.New("2006", "1", "2"),
			checks: []DateChecker{DateMustBeReal()},
		},
		"date must be real invalid": {
			input:    date.New("2000", "22", "2"),
			checks:   []DateChecker{DateMustBeReal()},
			expected: With(name, DateMustBeRealError{Label: label}),
		},
		"date must be past": {
			input:  date.New("2006", "1", "2"),
			checks: []DateChecker{DateMustBePast()},
		},
		"date must be past invalid": {
			input:    date.New("2222", "2", "2"),
			checks:   []DateChecker{DateMustBePast()},
			expected: With(name, DateMustBePastError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.Date(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}

func TestCheckAddress(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    *place.Address
		checks   []AddressChecker
		expected List
	}{
		"selected": {
			input:  &place.Address{},
			checks: []AddressChecker{Selected()},
		},
		"selected invalid": {
			input:    nil,
			checks:   []AddressChecker{Selected()},
			expected: With(name, SelectError{Label: label}),
		},
		"selected invalid custom": {
			input:    nil,
			checks:   []AddressChecker{Selected().CustomError()},
			expected: With(name, CustomError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.Address(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}

func TestCheckBool(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    bool
		checks   []BoolChecker
		expected List
	}{
		"selected": {
			input:  true,
			checks: []BoolChecker{Selected()},
		},
		"selected invalid": {
			input:    false,
			checks:   []BoolChecker{Selected()},
			expected: With(name, SelectError{Label: label}),
		},
		"selected invalid custom": {
			input:    false,
			checks:   []BoolChecker{Selected().CustomError()},
			expected: With(name, CustomError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.Bool(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}

func TestCheckOptions(t *testing.T) {
	name := "field-name"
	label := "translation-label"

	testcases := map[string]struct {
		input    []string
		checks   []OptionsChecker
		expected List
	}{
		"selected": {
			input:  []string{"a"},
			checks: []OptionsChecker{Selected()},
		},
		"selected invalid": {
			input:    []string{},
			checks:   []OptionsChecker{Selected()},
			expected: With(name, SelectError{Label: label}),
		},
		"selected invalid custom": {
			input:    []string{},
			checks:   []OptionsChecker{Selected().CustomError()},
			expected: With(name, CustomError{Label: label}),
		},
		"select": {
			input:  []string{"a", "b"},
			checks: []OptionsChecker{Select("a", "b", "c")},
		},
		"select invalid": {
			input:    []string{"a", "d"},
			checks:   []OptionsChecker{Select("a", "b", "c")},
			expected: With(name, SelectError{Label: label}),
		},
		"select invalid custom": {
			input:    []string{"a", "d"},
			checks:   []OptionsChecker{Select("a", "b", "c").CustomError()},
			expected: With(name, CustomError{Label: label}),
		},
	}

	for scenario, tc := range testcases {
		t.Run(scenario, func(t *testing.T) {
			var errors List

			errors.Options(name, label, tc.input, tc.checks...)

			assert.Equal(t, tc.expected, errors)
		})
	}
}
