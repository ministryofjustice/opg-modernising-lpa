package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/stretchr/testify/assert"
)

func TestPersonToNotifyFullName(t *testing.T) {
	assert.Equal(t, "First Last", PersonToNotify{FirstNames: "First", LastName: "Last"}.FullName())
}

func TestPeopleToNotifyGet(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPersonToNotify PersonToNotify
		id                     actoruid.UID
		expectedFound          bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}},
			expectedPersonToNotify: PersonToNotify{UID: uid1, FirstNames: "Bob"},
			id:                     uid1,
			expectedFound:          true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}},
			expectedPersonToNotify: PersonToNotify{},
			id:                     actoruid.New(),
			expectedFound:          false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.peopleToNotify.Get(tc.id)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedPersonToNotify, a)
		})
	}
}

func TestPeopleToNotifyPut(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()

	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPeopleToNotify PeopleToNotify
		updatedPersonToNotify  PersonToNotify
		expectedUpdated        bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{UID: uid1}, {UID: uid2}},
			expectedPeopleToNotify: PeopleToNotify{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}},
			updatedPersonToNotify:  PersonToNotify{UID: uid1, FirstNames: "Bob"},
			expectedUpdated:        true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{UID: uid1}, {UID: uid2}},
			expectedPeopleToNotify: PeopleToNotify{{UID: uid1}, {UID: uid2}},
			updatedPersonToNotify:  PersonToNotify{UID: uid3, FirstNames: "Bob"},
			expectedUpdated:        false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.peopleToNotify.Put(tc.updatedPersonToNotify)

			assert.Equal(t, tc.expectedUpdated, deleted)
			assert.Equal(t, tc.expectedPeopleToNotify, tc.peopleToNotify)
		})
	}
}

func TestPeopleToNotifyDelete(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()

	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPeopleToNotify PeopleToNotify
		personToNotifyToDelete PersonToNotify
		expectedDeleted        bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{UID: uid1}, {UID: uid2}},
			expectedPeopleToNotify: PeopleToNotify{{UID: uid1}},
			personToNotifyToDelete: PersonToNotify{UID: uid2},
			expectedDeleted:        true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{UID: uid1}, {UID: uid2}},
			expectedPeopleToNotify: PeopleToNotify{{UID: uid1}, {UID: uid2}},
			personToNotifyToDelete: PersonToNotify{UID: uid3},
			expectedDeleted:        false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.peopleToNotify.Delete(tc.personToNotifyToDelete)

			assert.Equal(t, tc.expectedDeleted, deleted)
			assert.Equal(t, tc.expectedPeopleToNotify, tc.peopleToNotify)
		})
	}
}
