package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Donor struct {
	UID                       actoruid.UID   `json:"uid"`
	FirstNames                string         `json:"firstNames"`
	LastName                  string         `json:"lastName"`
	Email                     string         `json:"email"`
	OtherNamesKnownBy         string         `json:"otherNamesKnownBy,omitempty"`
	DateOfBirth               date.Date      `json:"dateOfBirth"`
	Address                   place.Address  `json:"address"`
	ContactLanguagePreference localize.Lang  `json:"contactLanguagePreference"`
	IdentityCheck             *IdentityCheck `json:"identityCheck,omitempty"`

	// Mobile is only set for online donors who have provided one
	Mobile string `json:"mobile,omitempty"`

	Channel Channel `json:"-"`
}

func (d Donor) FullName() string {
	return d.FirstNames + " " + d.LastName
}
