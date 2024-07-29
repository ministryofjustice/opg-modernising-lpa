package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyFullName(t *testing.T) {
	assert.Equal(t, "First Last", Attorney{FirstNames: "First", LastName: "Last"}.FullName())
}

func TestAttorneyChannel(t *testing.T) {
	assert.Equal(t, ChannelOnline, Attorney{Email: "a@example.com"}.Channel())
	assert.Equal(t, ChannelPaper, Attorney{}.Channel())
}

func TestAttorneysLen(t *testing.T) {
	testcases := map[string]struct {
		attorneys Attorneys
		len       int
	}{
		"trust corporation": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
			len:       1,
		},
		"attorneys": {
			attorneys: Attorneys{Attorneys: []Attorney{{}, {}}},
			len:       2,
		},
		"both": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}, Attorneys: []Attorney{{}, {}}},
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
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "a", Address: place.Address{Line1: "a"}},
				Attorneys: []Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c", Address: place.Address{Line1: "c"}},
				},
			},
			expected: true,
		},
		"trust corporation incomplete": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "a"},
				Attorneys: []Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c", Address: place.Address{Line1: "c"}},
				},
			},
			expected: false,
		},
		"attorney incomplete": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "a", Address: place.Address{Line1: "a"}},
				Attorneys: []Attorney{
					{FirstNames: "b", Address: place.Address{Line1: "b"}},
					{FirstNames: "c"},
				},
			},
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
	attorneys := Attorneys{
		TrustCorporation: TrustCorporation{Address: place.Address{Line1: "a"}},
		Attorneys: []Attorney{
			{Address: place.Address{Line1: "b"}},
			{Address: place.Address{Line1: "a"}},
		},
	}

	assert.Equal(t, []place.Address{{Line1: "a"}, {Line1: "b"}}, attorneys.Addresses())
}

func TestAttorneysGet(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	testCases := map[string]struct {
		attorneys        Attorneys
		expectedAttorney Attorney
		uid              actoruid.UID
		expectedFound    bool
	}{
		"attorney exists": {
			attorneys:        Attorneys{Attorneys: []Attorney{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}}},
			expectedAttorney: Attorney{UID: uid1, FirstNames: "Bob"},
			uid:              uid1,
			expectedFound:    true,
		},
		"attorney does not exist": {
			attorneys:        Attorneys{Attorneys: []Attorney{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}}},
			expectedAttorney: Attorney{},
			uid:              actoruid.New(),
			expectedFound:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			a, found := tc.attorneys.Get(tc.uid)

			assert.Equal(t, tc.expectedFound, found)
			assert.Equal(t, tc.expectedAttorney, a)
		})
	}
}

func TestAttorneysPut(t *testing.T) {
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	newAttorney := Attorney{UID: actoruid.New(), FirstNames: "Bob"}

	testCases := map[string]struct {
		attorneys         Attorneys
		expectedAttorneys Attorneys
		updatedAttorney   Attorney
	}{
		"attorney exists": {
			attorneys:         Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			expectedAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid1, FirstNames: "Bob"}, {UID: uid2}}},
			updatedAttorney:   Attorney{UID: uid1, FirstNames: "Bob"},
		},
		"attorney does not exist": {
			attorneys:         Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			expectedAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}, newAttorney}},
			updatedAttorney:   newAttorney,
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
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	testCases := map[string]struct {
		attorneys         Attorneys
		expectedAttorneys Attorneys
		attorneyToDelete  Attorney
		expectedDeleted   bool
	}{
		"attorney exists": {
			attorneys:         Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			expectedAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}}},
			attorneyToDelete:  Attorney{UID: uid2},
			expectedDeleted:   true,
		},
		"attorney does not exist": {
			attorneys:         Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			expectedAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			attorneyToDelete:  Attorney{UID: actoruid.New()},
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

func TestAttorneysNames(t *testing.T) {
	testcases := map[string]struct {
		attorneys  Attorneys
		fullNames  []string
		firstNames []string
	}{
		"empty": {},
		"attorneys": {
			attorneys: Attorneys{
				Attorneys: []Attorney{
					{FirstNames: "Bob Alan George", LastName: "Jones"},
					{FirstNames: "Samantha", LastName: "Smith"},
					{FirstNames: "Abby Helen", LastName: "Burns-Simpson"},
				},
			},
			fullNames:  []string{"Bob Alan George Jones", "Samantha Smith", "Abby Helen Burns-Simpson"},
			firstNames: []string{"Bob Alan George", "Samantha", "Abby Helen"},
		},
		"trust corporation": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "Corp corp"},
			},
			fullNames:  []string{"Corp corp"},
			firstNames: []string{"Corp corp"},
		},
		"both": {
			attorneys: Attorneys{
				TrustCorporation: TrustCorporation{Name: "Corp corp"},
				Attorneys: []Attorney{
					{FirstNames: "Bob", LastName: "Jones"},
				},
			},
			fullNames:  []string{"Corp corp", "Bob Jones"},
			firstNames: []string{"Corp corp", "Bob"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.fullNames, tc.attorneys.FullNames())
			assert.Equal(t, tc.firstNames, tc.attorneys.FirstNames())
		})
	}
}

func TestTrustCorporationChannel(t *testing.T) {
	assert.Equal(t, ChannelOnline, TrustCorporation{Email: "a@example.com"}.Channel())
	assert.Equal(t, ChannelPaper, TrustCorporation{}.Channel())
}
