package validation

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

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
		"mobile invalid": {
			input:    "01152222222",
			checks:   []StringChecker{Mobile()},
			expected: With(name, MobileError{Label: label}),
		},
		"email": {
			input:  "name@example.com",
			checks: []StringChecker{Email()},
		},
		"email invalid": {
			input:    "example.com",
			checks:   []StringChecker{Email()},
			expected: With(name, EmailError{Label: label}),
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
			input:  date.FromParts("2006", "1", "2"),
			checks: []DateChecker{DateMissing()},
		},
		"date missing invalid": {
			input:    date.Date{Day: "10"},
			checks:   []DateChecker{DateMissing()},
			expected: With(name, DateMissingError{Label: label, MissingMonth: true, MissingYear: true}),
		},
		"date missing invalid all": {
			input:    date.Date{},
			checks:   []DateChecker{DateMissing()},
			expected: With(name, EnterError{Label: label}),
		},
		"date must be real": {
			input:  date.FromParts("2006", "1", "2"),
			checks: []DateChecker{DateMustBeReal()},
		},
		"date must be real invalid": {
			input:    date.FromParts("2000", "22", "2"),
			checks:   []DateChecker{DateMustBeReal()},
			expected: With(name, DateMustBeRealError{Label: label}),
		},
		"date must be past": {
			input:  date.FromParts("2006", "1", "2"),
			checks: []DateChecker{DateMustBePast()},
		},
		"date must be past invalid": {
			input:    date.FromParts("2222", "2", "2"),
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
		"select": {
			input:  []string{"a", "b"},
			checks: []OptionsChecker{Select("a", "b", "c")},
		},
		"select invalid": {
			input:    []string{"a", "d"},
			checks:   []OptionsChecker{Select("a", "b", "c")},
			expected: With(name, SelectError{Label: label}),
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
