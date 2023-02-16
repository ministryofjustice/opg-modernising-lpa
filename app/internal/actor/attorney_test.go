package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttorneysGet(t *testing.T) {
	testCases := map[string]struct {
		attorneys        Attorneys
		expectedAttorney Attorney
		id               string
		expectedFound    bool
	}{
		"attorney exists": {
			attorneys:        Attorneys{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			expectedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			id:               "1",
			expectedFound:    true,
		},
		"attorney does not exist": {
			attorneys:        Attorneys{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			expectedAttorney: Attorney{},
			id:               "4",
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.attorneys.Get(tc.id)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestAttorneysPut(t *testing.T) {
	testCases := map[string]struct {
		attorneys         Attorneys
		expectedAttorneys Attorneys
		updatedAttorney   Attorney
		expectedUpdated   bool
	}{
		"attorney exists": {
			attorneys:         Attorneys{{ID: "1"}, {ID: "2"}},
			expectedAttorneys: Attorneys{{ID: "1", FirstNames: "Bob"}, {ID: "2"}},
			updatedAttorney:   Attorney{ID: "1", FirstNames: "Bob"},
			expectedUpdated:   true,
		},
		"attorney does not exist": {
			attorneys:         Attorneys{{ID: "1"}, {ID: "2"}},
			expectedAttorneys: Attorneys{{ID: "1"}, {ID: "2"}},
			updatedAttorney:   Attorney{ID: "3", FirstNames: "Bob"},
			expectedUpdated:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.attorneys.Put(tc.updatedAttorney)

			assert.Equal(t, tc.expectedUpdated, deleted)
			assert.Equal(t, tc.expectedAttorneys, tc.attorneys)
		})
	}
}

func TestAttorneysDelete(t *testing.T) {
	testCases := map[string]struct {
		attorneys         Attorneys
		expectedAttorneys Attorneys
		attorneyToDelete  Attorney
		expectedDeleted   bool
	}{
		"attorney exists": {
			attorneys:         Attorneys{{ID: "1"}, {ID: "2"}},
			expectedAttorneys: Attorneys{{ID: "1"}},
			attorneyToDelete:  Attorney{ID: "2"},
			expectedDeleted:   true,
		},
		"attorney does not exist": {
			attorneys:         Attorneys{{ID: "1"}, {ID: "2"}},
			expectedAttorneys: Attorneys{{ID: "1"}, {ID: "2"}},
			attorneyToDelete:  Attorney{ID: "3"},
			expectedDeleted:   false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			deleted := tc.attorneys.Delete(tc.attorneyToDelete)

			assert.Equal(t, tc.expectedDeleted, deleted)
			assert.Equal(t, tc.expectedAttorneys, tc.attorneys)
		})
	}
}

func TestAttorneysFullNames(t *testing.T) {
	attorneys := Attorneys{
		{
			FirstNames: "Bob Alan George",
			LastName:   "Jones",
		},
		{
			FirstNames: "Samantha",
			LastName:   "Smith",
		},
		{
			FirstNames: "Abby Helen",
			LastName:   "Burns-Simpson",
		},
	}

	assert.Equal(t, "Bob Alan George Jones, Samantha Smith and Abby Helen Burns-Simpson", attorneys.FullNames())
}

func TestAttorneysFirstNames(t *testing.T) {
	attorneys := Attorneys{
		{
			FirstNames: "Bob Alan George",
			LastName:   "Jones",
		},
		{
			FirstNames: "Samantha",
			LastName:   "Smith",
		},
		{
			FirstNames: "Abby Helen",
			LastName:   "Burns-Simpson",
		},
	}

	assert.Equal(t, "Bob Alan George, Samantha and Abby Helen", attorneys.FirstNames())
}

func TestConcatSentence(t *testing.T) {
	assert.Equal(t, "Bob Smith, Alice Jones, John Doe and Paul Compton", concatSentence([]string{"Bob Smith", "Alice Jones", "John Doe", "Paul Compton"}))
	assert.Equal(t, "Bob Smith, Alice Jones and John Doe", concatSentence([]string{"Bob Smith", "Alice Jones", "John Doe"}))
	assert.Equal(t, "Bob Smith and John Doe", concatSentence([]string{"Bob Smith", "John Doe"}))
	assert.Equal(t, "Bob Smith", concatSentence([]string{"Bob Smith"}))
	assert.Equal(t, "", concatSentence([]string{}))
}
