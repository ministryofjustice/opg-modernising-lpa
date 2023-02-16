package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeopleToNotifyGet(t *testing.T) {
	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPersonToNotify PersonToNotify
		id                     string
		expectedFound          bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			expectedPersonToNotify: PersonToNotify{ID: "1", FirstNames: "Bob"},
			id:                     "1",
			expectedFound:          true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			expectedPersonToNotify: PersonToNotify{},
			id:                     "4",
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
	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPeopleToNotify PeopleToNotify
		updatedPersonToNotify  PersonToNotify
		expectedUpdated        bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{ID: "1"}, {ID: "2"}},
			expectedPeopleToNotify: PeopleToNotify{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			updatedPersonToNotify:  PersonToNotify{ID: "1", FirstNames: "Bob"},
			expectedUpdated:        true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{ID: "1"}, {ID: "2"}},
			expectedPeopleToNotify: PeopleToNotify{{ID: "1"}, {ID: "2"}},
			updatedPersonToNotify:  PersonToNotify{ID: "3", FirstNames: "Bob"},
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
	testCases := map[string]struct {
		peopleToNotify         PeopleToNotify
		expectedPeopleToNotify PeopleToNotify
		personToNotifyToDelete PersonToNotify
		expectedDeleted        bool
	}{
		"personToNotify exists": {
			peopleToNotify:         PeopleToNotify{{ID: "1"}, {ID: "2"}},
			expectedPeopleToNotify: PeopleToNotify{{ID: "1"}},
			personToNotifyToDelete: PersonToNotify{ID: "2"},
			expectedDeleted:        true,
		},
		"personToNotify does not exist": {
			peopleToNotify:         PeopleToNotify{{ID: "1"}, {ID: "2"}},
			expectedPeopleToNotify: PeopleToNotify{{ID: "1"}, {ID: "2"}},
			personToNotifyToDelete: PersonToNotify{ID: "3"},
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
