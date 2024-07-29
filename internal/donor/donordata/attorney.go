package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Attorney contains details about an attorney or replacement attorney, provided by the applicant
type Attorney struct {
	// UID for the actor
	UID actoruid.UID
	// First names of the attorney
	FirstNames string
	// Last name of the attorney
	LastName string
	// Email of the attorney
	Email string
	// Date of birth of the attorney
	DateOfBirth date.Date
	// Address of the attorney
	Address place.Address
}

func (a Attorney) FullName() string {
	return fmt.Sprintf("%s %s", a.FirstNames, a.LastName)
}

func (a Attorney) Channel() Channel {
	if a.Email != "" {
		return ChannelOnline
	}

	return ChannelPaper
}
