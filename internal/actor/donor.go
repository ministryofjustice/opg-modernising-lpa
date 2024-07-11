package actor

import (
	"fmt"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Donor contains details about the donor, provided by the applicant
type Donor struct {
	// UID for the actor
	UID actoruid.UID
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
	// Channel is how the Donor is applying for their LPA (paper or online)
	Channel Channel
	// ContactLanguagePreference is the language the donor prefers to receive notifications in
	ContactLanguagePreference localize.Lang
	// LpaLanguagePreference is the language the donor prefers to receive the registered LPA in
	LpaLanguagePreference localize.Lang
}

func (d Donor) FullName() string {
	return strings.Trim(fmt.Sprintf("%s %s", d.FirstNames, d.LastName), " ")
}

// AuthorisedSignatory contains details of the person who will sign the LPA on the donor's behalf
type AuthorisedSignatory struct {
	FirstNames string
	LastName   string
}

func (s AuthorisedSignatory) FullName() string {
	return fmt.Sprintf("%s %s", s.FirstNames, s.LastName)
}

// IndependentWitness contains details of the person who will also witness the signing of the LPA
type IndependentWitness struct {
	FirstNames     string
	LastName       string
	HasNonUKMobile bool
	Mobile         string
	Address        place.Address
}

func (w IndependentWitness) FullName() string {
	return fmt.Sprintf("%s %s", w.FirstNames, w.LastName)
}
