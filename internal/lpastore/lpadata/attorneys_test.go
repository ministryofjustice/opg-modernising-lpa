package lpadata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/stretchr/testify/assert"
)

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

func TestAttorneysFullNames(t *testing.T) {
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
		})
	}
}
