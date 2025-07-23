package actor

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Actor struct {
	Type       Type
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Address    place.Address
}

func (a Actor) FullName() string {
	return a.FirstNames + " " + a.LastName
}
