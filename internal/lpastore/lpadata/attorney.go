package lpadata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Attorney struct {
	UID                       actoruid.UID
	FirstNames                string
	LastName                  string
	DateOfBirth               date.Date
	Email                     string
	Address                   place.Address
	Mobile                    string
	SignedAt                  time.Time
	ContactLanguagePreference localize.Lang
	Channel                   Channel
	Removed                   bool
}

func (a Attorney) FullName() string {
	return a.FirstNames + " " + a.LastName
}
