package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"

type Actor struct {
	Type       Type
	UID        actoruid.UID
	FirstNames string
	LastName   string
}

func (a Actor) FullName() string {
	return a.FirstNames + " " + a.LastName
}
