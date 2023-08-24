package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Donor contains details about the donor, provided by the applicant
type Donor struct {
	// First names of the donor
	FirstNames string
	// Last name of the donor
	LastName string
	// Email of the donor
	Email string
	// Other names the donor is known by
	OtherNames string
	// Date of birth of the donor
	DateOfBirth date.Date
	// Address of the donor
	Address place.Address
}

func (d Donor) FullName() string {
	return fmt.Sprintf("%s %s", d.FirstNames, d.LastName)
}
