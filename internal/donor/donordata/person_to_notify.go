package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// PersonToNotify contains details about a person to notify, provided by the applicant
type PersonToNotify struct {
	UID actoruid.UID
	// First names of the person to notify
	FirstNames string
	// Last name of the person to notify
	LastName string
	// Address of the person to notify
	Address place.Address
}

func (p PersonToNotify) FullName() string {
	return p.FirstNames + " " + p.LastName
}
