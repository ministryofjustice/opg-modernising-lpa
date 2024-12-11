package lpadata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"

type AuthorisedSignatory struct {
	UID        actoruid.UID `json:"uid"`
	FirstNames string       `json:"firstNames"`
	LastName   string       `json:"lastName"`
}

func (a AuthorisedSignatory) FullName() string {
	return a.FirstNames + " " + a.LastName
}
