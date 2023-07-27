package actor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyFullName(t *testing.T) {
	assert.Equal(t, "First Last", Attorney{FirstNames: "First", LastName: "Last"}.FullName())
}

func TestAttorneysLen(t *testing.T) {
	testcases := map[string]struct {
		attorneys Attorneys
		len       int
	}{
		"trust corporation": {
			attorneys: NewAttorneys(&TrustCorporation{}, nil),
			len:       1,
		},
		"attorneys": {
			attorneys: NewAttorneys(nil, []Attorney{{}, {}}),
			len:       2,
		},
		"both": {
			attorneys: NewAttorneys(&TrustCorporation{}, []Attorney{{}, {}}),
			len:       3,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.len, tc.attorneys.Len())
		})
	}
}

func TestAttorneysComplete(t *testing.T) {
	testcases := map[string]struct {
		attorneys Attorneys
		expected  bool
	}{
		"complete": {
			attorneys: NewAttorneys(
				&TrustCorporation{Name: "a", Address: place.Address{Line1: "a"}},
				[]Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c", Address: place.Address{Line1: "c"}}}),
			expected: true,
		},
		"trust corporation incomplete": {
			attorneys: NewAttorneys(
				&TrustCorporation{Name: "a"},
				[]Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c", Address: place.Address{Line1: "c"}}}),
			expected: false,
		},
		"attorney incomplete": {
			attorneys: NewAttorneys(
				&TrustCorporation{Name: "a", Address: place.Address{Line1: "a"}},
				[]Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c"}}),
			expected: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.attorneys.Complete())
		})
	}
}

func TestAttorneysAddresses(t *testing.T) {
	attorneys := NewAttorneys(
		&TrustCorporation{Address: place.Address{Line1: "a"}},
		[]Attorney{
			{Address: place.Address{Line1: "b"}},
			{Address: place.Address{Line1: "a"}},
		})

	assert.Equal(t, []place.Address{{Line1: "a"}, {Line1: "b"}}, attorneys.Addresses())
}

func TestAttorneysTrustCorporation(t *testing.T) {
	trustCorporation := TrustCorporation{Name: "a"}
	attorneys := NewAttorneys(&trustCorporation, nil)

	tc, ok := attorneys.TrustCorporation()
	assert.True(t, ok)
	assert.Equal(t, trustCorporation, tc)
}

func TestAttorneysTrustCorporationMissing(t *testing.T) {
	attorneys := NewAttorneys(nil, nil)

	tc, ok := attorneys.TrustCorporation()
	assert.False(t, ok)
	assert.Equal(t, TrustCorporation{}, tc)
}

func TestAttorneysSetTrustCorporation(t *testing.T) {
	trustCorporation := TrustCorporation{Name: "a"}
	attorneys := NewAttorneys(nil, nil)
	attorneys.SetTrustCorporation(trustCorporation)

	tc, ok := attorneys.TrustCorporation()
	assert.True(t, ok)
	assert.Equal(t, trustCorporation, tc)
}

func TestAttorneysAttorneys(t *testing.T) {
	expected := []Attorney{{FirstNames: "a"}}
	attorneys := NewAttorneys(&TrustCorporation{}, expected)

	assert.Equal(t, expected, attorneys.Attorneys())
}

func TestAttorneysGet(t *testing.T) {
	testCases := map[string]struct {
		attorneys        Attorneys
		expectedAttorney Attorney
		id               string
		expectedFound    bool
	}{
		"attorney exists": {
			attorneys:        NewAttorneys(nil, []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}}),
			expectedAttorney: Attorney{ID: "1", FirstNames: "Bob"},
			id:               "1",
			expectedFound:    true,
		},
		"attorney does not exist": {
			attorneys:        NewAttorneys(nil, []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}}),
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
	}{
		"attorney exists": {
			attorneys:         NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}}),
			expectedAttorneys: NewAttorneys(nil, []Attorney{{ID: "1", FirstNames: "Bob"}, {ID: "2"}}),
			updatedAttorney:   Attorney{ID: "1", FirstNames: "Bob"},
		},
		"attorney does not exist": {
			attorneys:         NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}}),
			expectedAttorneys: NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}, {ID: "3", FirstNames: "Bob"}}),
			updatedAttorney:   Attorney{ID: "3", FirstNames: "Bob"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tc.attorneys.Put(tc.updatedAttorney)

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
			attorneys:         NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}}),
			expectedAttorneys: NewAttorneys(nil, []Attorney{{ID: "1"}}),
			attorneyToDelete:  Attorney{ID: "2"},
			expectedDeleted:   true,
		},
		"attorney does not exist": {
			attorneys:         NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}}),
			expectedAttorneys: NewAttorneys(nil, []Attorney{{ID: "1"}, {ID: "2"}}),
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
	attorneys := Attorneys{attorneys: []Attorney{
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
	}}

	assert.Equal(t, []string{"Bob Alan George Jones", "Samantha Smith", "Abby Helen Burns-Simpson"}, attorneys.FullNames())
}

func TestAttorneysFirstNames(t *testing.T) {
	attorneys := Attorneys{attorneys: []Attorney{
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
	}}

	assert.Equal(t, []string{"Bob Alan George", "Samantha", "Abby Helen"}, attorneys.FirstNames())
}
