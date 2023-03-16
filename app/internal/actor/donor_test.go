package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDonorFullName(t *testing.T) {
	d := Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", d.FullName())
}

func TestDonorPossessiveFullName(t *testing.T) {
	testCases := map[string]struct {
		Donor    Donor
		Expected string
	}{
		"not ending in s": {
			Donor:    Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"},
			Expected: "Bob Alan George Smith Jones-Doe’s",
		},
		"ending in s": {
			Donor:    Donor{FirstNames: "Bob Alan George", LastName: "Smith Jones"},
			Expected: "Bob Alan George Smith Jones’",
		},
		"no lastname defaults to firstname - not ending in s": {
			Donor:    Donor{FirstNames: "Bob Alan George"},
			Expected: "Bob Alan George’s",
		},
		"no lastname defaults to firstname - ending in s": {
			Donor:    Donor{FirstNames: "Bob Alan James"},
			Expected: "Bob Alan James’",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.Donor.PossessiveFullName())
		})
	}
}
