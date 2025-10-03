package donordata

import (
	"fmt"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Donor contains details about the donor, provided by the applicant
type Donor struct {
	// UID for the actor
	UID actoruid.UID `stubhash:"-"`
	// First names of the donor
	FirstNames string `relatedhash:"-"`
	// Last name of the donor
	LastName string
	// Email of the donor
	Email string `checkhash:"-" relatedhash:"-" stubhash:"-"`
	// Other names the donor is known by
	OtherNames string `relatedhash:"-" stubhash:"-"`
	// Date of birth of the donor
	DateOfBirth date.Date `relatedhash:"-"`
	// Address of the donor
	Address place.Address
	// International address of the donor
	InternationalAddress place.InternationalAddress `relatedhash:"-" stubhash:"-"`
	// Mobile phone number to contact the donor
	Mobile string `checkhash:"-" relatedhash:"-" stubhash:"-"`
	// ThinksCanSign is what the donor thinks about their ability to sign online
	ThinksCanSign YesNoMaybe `relatedhash:"-" stubhash:"-"`
	// CanSign is Yes if the donor has said they will sign online
	CanSign form.YesNo `relatedhash:"-" stubhash:"-"`
	// Channel is how the Donor is applying for their LPA (paper or online)
	Channel lpadata.Channel `relatedhash:"-" stubhash:"-"`
	// ContactLanguagePreference is the language the donor prefers to receive notifications in
	ContactLanguagePreference localize.Lang `relatedhash:"-" stubhash:"-"`
	// LpaLanguagePreference is the language the donor prefers to receive the registered LPA in
	LpaLanguagePreference localize.Lang `checkhash:"-" relatedhash:"-" stubhash:"-"`
}

func (d Donor) FullName() string {
	return strings.Trim(fmt.Sprintf("%s %s", d.FirstNames, d.LastName), " ")
}

func (d Donor) NameHasChanged(firstNames, lastName, otherNames string) bool {
	return d.FirstNames != firstNames || d.LastName != lastName || d.OtherNames != otherNames
}

func (d Donor) IsUnder18() bool {
	return !d.Is18On(time.Now())
}

func (d Donor) Is18On(t time.Time) bool {
	return t.After(d.DateOfBirth.AddDate(18, 0, 0).Time())
}
