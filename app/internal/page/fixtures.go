package page

import (
	"time"

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
		DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
	}
}

func MakeAttorney(firstNames string) Attorney {
	return Attorney{
		ID:          firstNames + "Smith",
		FirstNames:  firstNames,
		LastName:    "Smith",
		Email:       firstNames + "@example.org",
		DateOfBirth: time.Date(2000, time.January, 2, 3, 4, 5, 6, time.UTC),
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
		DateOfBirth:             time.Date(1997, time.January, 2, 3, 4, 5, 6, time.UTC),
		Relationship:            "friend",
		RelationshipDescription: "",
		RelationshipLength:      "gte-2-years",
	}
}
