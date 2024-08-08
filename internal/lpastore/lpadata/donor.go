package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Donor struct {
	UID                       actoruid.UID
	FirstNames                string
	LastName                  string
	Email                     string
	OtherNames                string
	DateOfBirth               date.Date
	Address                   place.Address
	Channel                   Channel
	ContactLanguagePreference localize.Lang
	IdentityCheck             IdentityCheck
}

func (d Donor) FullName() string {
	return d.FirstNames + " " + d.LastName
}
