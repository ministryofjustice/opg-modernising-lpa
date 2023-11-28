package actor

import (
	"slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// PersonToNotify contains details about a person to notify, provided by the applicant
type PersonToNotify struct {
	// First names of the person to notify
	FirstNames string
	// Last name of the person to notify
	LastName string
	// Address of the person to notify
	Address place.Address
	// Identifies the person to notify being edited
	ID string
}

func (p PersonToNotify) FullName() string {
	return p.FirstNames + " " + p.LastName
}

type PeopleToNotify []PersonToNotify

func (ps PeopleToNotify) Get(id string) (PersonToNotify, bool) {
	idx := slices.IndexFunc(ps, func(p PersonToNotify) bool { return p.ID == id })
	if idx == -1 {
		return PersonToNotify{}, false
	}

	return ps[idx], true
}

func (ps PeopleToNotify) Put(person PersonToNotify) bool {
	idx := slices.IndexFunc(ps, func(p PersonToNotify) bool { return p.ID == person.ID })
	if idx == -1 {
		return false
	}

	ps[idx] = person
	return true
}

func (ps *PeopleToNotify) Delete(personToNotify PersonToNotify) bool {
	idx := slices.IndexFunc(*ps, func(p PersonToNotify) bool { return p.ID == personToNotify.ID })
	if idx == -1 {
		return false
	}

	*ps = slices.Delete(*ps, idx, idx+1)

	return true
}
