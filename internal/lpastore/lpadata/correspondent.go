package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Correspondent struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Email      string
	Phone      string
	Address    place.Address
}

func (c Correspondent) FullName() string {
	return c.FirstNames + " " + c.LastName
}
