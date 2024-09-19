package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type IndependentWitness struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Mobile     string
	Address    place.Address
}

func (w IndependentWitness) FullName() string {
	return w.FirstNames + " " + w.LastName
}
