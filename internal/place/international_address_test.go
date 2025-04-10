package place

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInternationalAddressToAddress(t *testing.T) {
	testcases := map[string]struct {
		in  InternationalAddress
		out Address
	}{
		"only building number": {
			in: InternationalAddress{
				BuildingNumber: "123",
				StreetName:     "Cool St",
				Town:           "Cooltown",
				Region:         "Coolshire",
				PostalCode:     "ABC",
				Country:        "What",
			},
			out: Address{
				Line1:      "123 Cool St",
				Line2:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"only building name": {
			in: InternationalAddress{
				BuildingName: "Cool Building",
				StreetName:   "Cool St",
				Town:         "Cooltown",
				Region:       "Coolshire",
				PostalCode:   "ABC",
				Country:      "What",
			},
			out: Address{
				Line1:      "Cool Building",
				Line2:      "Cool St",
				Line3:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"only apartment number": {
			in: InternationalAddress{
				ApartmentNumber: "Flat 123a",
				StreetName:      "Cool St",
				Town:            "Cooltown",
				Region:          "Coolshire",
				PostalCode:      "ABC",
				Country:         "What",
			},
			out: Address{
				Line1:      "Flat 123a, Cool St",
				Line2:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"apartment number and building number": {
			in: InternationalAddress{
				ApartmentNumber: "Flat 123a",
				BuildingNumber:  "5",
				StreetName:      "Cool St",
				Town:            "Cooltown",
				Region:          "Coolshire",
				PostalCode:      "ABC",
				Country:         "What",
			},
			out: Address{
				Line1:      "Flat 123a, 5 Cool St",
				Line2:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"apartment number and building name": {
			in: InternationalAddress{
				ApartmentNumber: "Flat 123a",
				BuildingName:    "Flathouse",
				StreetName:      "Cool St",
				Town:            "Cooltown",
				Region:          "Coolshire",
				PostalCode:      "ABC",
				Country:         "What",
			},
			out: Address{
				Line1:      "Flat 123a, Flathouse",
				Line2:      "Cool St",
				Line3:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"building number and building name": {
			in: InternationalAddress{
				BuildingNumber: "5",
				BuildingName:   "Flathouse",
				StreetName:     "Cool St",
				Town:           "Cooltown",
				Region:         "Coolshire",
				PostalCode:     "ABC",
				Country:        "What",
			},
			out: Address{
				Line1:      "Flathouse",
				Line2:      "5 Cool St",
				Line3:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
		"all": {
			in: InternationalAddress{
				ApartmentNumber: "Flat 123a",
				BuildingNumber:  "5",
				BuildingName:    "Flathouse",
				StreetName:      "Cool St",
				Town:            "Cooltown",
				Region:          "Coolshire",
				PostalCode:      "ABC",
				Country:         "What",
			},
			out: Address{
				Line1:      "Flat 123a, Flathouse",
				Line2:      "5 Cool St",
				Line3:      "Cooltown",
				TownOrCity: "Coolshire",
				Postcode:   "ABC",
				Country:    "What",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.out, tc.in.ToAddress())
		})
	}
}
