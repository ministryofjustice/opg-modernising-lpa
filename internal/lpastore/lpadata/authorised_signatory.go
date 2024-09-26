package lpadata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"

type AuthorisedSignatory struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
}

func (a AuthorisedSignatory) FullName() string {
	return a.FirstNames + " " + a.LastName
}
