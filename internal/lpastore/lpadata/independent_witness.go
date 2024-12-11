package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type IndependentWitness struct {
	UID        actoruid.UID  `json:"uid"`
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Phone      string        `json:"phone"`
	Address    place.Address `json:"address"`
}

func (w IndependentWitness) FullName() string {
	return w.FirstNames + " " + w.LastName
}
