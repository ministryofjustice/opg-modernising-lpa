package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// IndependentWitness contains details of the person who will also witness the signing of the LPA
type IndependentWitness struct {
	FirstNames     string
	LastName       string
	HasNonUKMobile bool
	Mobile         string
	Address        place.Address
}

func (w IndependentWitness) FullName() string {
	return fmt.Sprintf("%s %s", w.FirstNames, w.LastName)
}