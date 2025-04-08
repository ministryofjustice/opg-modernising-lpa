package place

import "strings"

type InternationalAddress struct {
	ApartmentNumber string `json:"apartmentNumber"`
	BuildingNumber  string `json:"buildingNumber"`
	BuildingName    string `json:"buildingName"`
	StreetName      string `json:"streetName"`
	Town            string `json:"town"`
	Region          string `json:"region"`
	PostalCode      string `json:"postalCode"`
	Country         string `json:"country"`
}

func (a InternationalAddress) ToAddress() Address {
	// this is written to match TransformToAddress, with ApartmentNumber being
	// equivalent to SubBuildingName
	var line1, line2, line3 string
	if a.BuildingName != "" {
		if a.ApartmentNumber != "" {
			line1 = strings.TrimSpace(a.ApartmentNumber + ", " + a.BuildingName)
		} else {
			line1 = a.BuildingName
		}

		if a.BuildingNumber != "" {
			line2 = strings.TrimSpace(a.BuildingNumber + " " + a.StreetName)
		} else {
			line2 = a.StreetName
		}

		line3 = a.Town
	} else {
		line1 = strings.TrimSpace(a.BuildingNumber + " " + a.StreetName)
		line2 = a.Town
	}

	return Address{
		Line1:      line1,
		Line2:      line2,
		Line3:      line3,
		TownOrCity: a.Region, // or don't ask for region and put Town here instead of line2/3???
		Postcode:   a.PostalCode,
		Country:    a.Country,
	}
}
