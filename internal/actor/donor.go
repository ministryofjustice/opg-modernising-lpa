package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
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
	// ThinksCanSign is what the donor thinks about their ability to sign online
	ThinksCanSign YesNoMaybe
	// CanSign is Yes if the donor has said they will sign online
	CanSign form.YesNo
}

func (d Donor) FullName() string {
	return fmt.Sprintf("%s %s", d.FirstNames, d.LastName)
}

// Signatory contains details of the person who will sign the LPA on the donor's behalf
type Signatory struct {
	FirstNames string
	LastName   string
}

// IndependentWitness contains details of the person who will also witness the signing of the LPA
type IndependentWitness struct {
	FirstNames string
	LastName   string
	Mobile     string
}
