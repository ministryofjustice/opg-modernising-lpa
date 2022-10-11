package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/ordnance_survey"

	"github.com/stretchr/testify/assert"
)

func TestReadDate(t *testing.T) {
	date := readDate(time.Date(2020, time.March, 12, 0, 0, 0, 0, time.Local))

	assert.Equal(t, Date{Day: "12", Month: "3", Year: "2020"}, date)
}

func TestTransformAddressDetailsToAddress(t *testing.T) {
	testCases := []struct {
		name   string
		ad     ordnance_survey.AddressDetails
		wanted Address
	}{
		{
			"Building number no building name",
			ordnance_survey.AddressDetails{
				Address:           "1, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "1",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "1 MELTON ROAD", Line2: "", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		{
			"Building name no building number",
			ordnance_survey.AddressDetails{
				Address:           "1A, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "1A",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "1A", Line2: "MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		{
			"Building name and building number",
			ordnance_survey.AddressDetails{
				Address:           "MELTON HOUSE, 2 MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "2",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "MELTON HOUSE", Line2: "2 MELTON ROAD", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		{
			"Dependent Locality building number",
			ordnance_survey.AddressDetails{
				Address:           "3, MELTON ROAD, BIRMINGHAM, B14 7ET",
				BuildingName:      "",
				BuildingNumber:    "3",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "3 MELTON ROAD", Line2: "KINGS HEATH", Line3: "", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		{
			"Dependent Locality building name",
			ordnance_survey.AddressDetails{
				Address:           "MELTON HOUSE, MELTON ROAD, KINGS HEATH, BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "MELTON HOUSE", Line2: "MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
		{
			"Dependent Locality building name and building number",
			ordnance_survey.AddressDetails{
				Address:           "MELTON HOUSE, 5 MELTON ROAD, KINGS HEATH BIRMINGHAM, B14 7ET",
				BuildingName:      "MELTON HOUSE",
				BuildingNumber:    "5",
				ThoroughFareName:  "MELTON ROAD",
				DependentLocality: "KINGS HEATH",
				Town:              "BIRMINGHAM",
				Postcode:          "B14 7ET",
			},
			Address{Line1: "MELTON HOUSE", Line2: "5 MELTON ROAD", Line3: "KINGS HEATH", TownOrCity: "BIRMINGHAM", Postcode: "B14 7ET"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wanted, TransformAddressDetailsToAddress(tc.ad))
		})
	}
}
