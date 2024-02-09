package actor

import (
	"slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// PersonToNotify contains details about a person to notify, provided by the applicant
type PersonToNotify struct {
	UID UID
	// First names of the person to notify
	FirstNames string
	// Last name of the person to notify
	LastName string
	// Address of the person to notify
	Address place.Address
}

func (p PersonToNotify) FullName() string {
	return p.FirstNames + " " + p.LastName
}

type PeopleToNotify []PersonToNotify

func (ps PeopleToNotify) Get(uid UID) (PersonToNotify, bool) {
	idx := slices.IndexFunc(ps, func(p PersonToNotify) bool { return p.UID == uid })
	if idx == -1 {
		return PersonToNotify{}, false
	}

	return ps[idx], true
}

func (ps PeopleToNotify) Put(person PersonToNotify) bool {
	idx := slices.IndexFunc(ps, func(p PersonToNotify) bool { return p.UID == person.UID })
	if idx == -1 {
		return false
	}

	ps[idx] = person
	return true
}

func (ps *PeopleToNotify) Delete(personToNotify PersonToNotify) bool {
	idx := slices.IndexFunc(*ps, func(p PersonToNotify) bool { return p.UID == personToNotify.UID })
	if idx == -1 {
		return false
	}

	*ps = slices.Delete(*ps, idx, idx+1)

	return true
}
