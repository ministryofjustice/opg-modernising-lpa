package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// IndependentWitness contains details of the person who will also witness the signing of the LPA
type IndependentWitness struct {
	UID            actoruid.UID
	FirstNames     string
	LastName       string
	HasNonUKMobile bool   `checkhash:"-"`
	Mobile         string `checkhash:"-"`
	Address        place.Address
}

func (w IndependentWitness) FullName() string {
	return fmt.Sprintf("%s %s", w.FirstNames, w.LastName)
}
