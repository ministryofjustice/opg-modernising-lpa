package page

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

func MakePerson() Person {
	return Person{
		FirstNames: "Jose",
		LastName:   "Smith",
		Address: place.Address{
			Line1:      "1 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
		Email:       "simulate-delivered@notifications.service.gov.uk",
		DateOfBirth: date.New("2000", "1", "2"),
	}
}

func MakeAttorney(firstNames string) Attorney {
	return Attorney{
		ID:          firstNames + "Smith",
		FirstNames:  firstNames,
		LastName:    "Smith",
		Email:       firstNames + "@example.org",
		DateOfBirth: date.New("2000", "1", "2"),
		Address: place.Address{
			Line1:      "2 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
	}
}

func MakePersonToNotify(firstNames string) PersonToNotify {
	return PersonToNotify{
		ID:         firstNames + "Smith",
		FirstNames: firstNames,
		LastName:   "Smith",
		Email:      firstNames + "@example.org",
		Address: place.Address{
			Line1:      "4 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
	}
}

func MakeCertificateProvider(firstNames string) CertificateProvider {
	return CertificateProvider{
		FirstNames:              firstNames,
		LastName:                "Smith",
		Email:                   firstNames + "@example.org",
		Mobile:                  "07535111111",
		DateOfBirth:             date.New("1997", "1", "2"),
		Relationship:            "friend",
		RelationshipDescription: "",
		RelationshipLength:      "gte-2-years",
	}
}
