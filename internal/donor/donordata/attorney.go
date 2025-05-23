package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Attorney contains details about an attorney or replacement attorney, provided by the applicant
type Attorney struct {
	// UID for the actor
	UID actoruid.UID
	// First names of the attorney
	FirstNames string `relatedhash:"-"`
	// Last name of the attorney
	LastName string
	// Email of the attorney
	Email string `relatedhash:"-"`
	// Date of birth of the attorney
	DateOfBirth date.Date `relatedhash:"-"`
	// Address of the attorney
	Address place.Address
}

func (a Attorney) FullName() string {
	return fmt.Sprintf("%s %s", a.FirstNames, a.LastName)
}

func (a Attorney) Channel() lpadata.Channel {
	if a.Email != "" {
		return lpadata.ChannelOnline
	}

	return lpadata.ChannelPaper
}

func (a Attorney) NameHasChanged(firstNames, lastName string) bool {
	return a.FirstNames != firstNames || a.LastName != lastName
}
