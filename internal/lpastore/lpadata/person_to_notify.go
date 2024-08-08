package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type PersonToNotify struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Address    place.Address
}

func (p PersonToNotify) FullName() string {
	return p.FirstNames + " " + p.LastName
}
