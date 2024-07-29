package donordata

import (
	"slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

type PeopleToNotify []PersonToNotify

func (ps PeopleToNotify) Get(uid actoruid.UID) (PersonToNotify, bool) {
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
