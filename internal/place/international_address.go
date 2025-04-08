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
		if a.ApartmentNumber != "" {
			line1 = a.ApartmentNumber + ", " + line1
		}
		line2 = a.Town
	}

	return Address{
		Line1: line1,
		Line2: line2,
		Line3: line3,
		// putting region here is weird, but don't really have a better place to
		// store in the current structure
		TownOrCity: a.Region,
		Postcode:   a.PostalCode,
		Country:    a.Country,
	}
}
