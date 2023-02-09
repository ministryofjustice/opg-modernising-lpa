package actor

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"golang.org/x/exp/slices"
)

type PersonToNotify struct {
	FirstNames string
	LastName   string
	Email      string
	Address    place.Address
	ID         string
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
