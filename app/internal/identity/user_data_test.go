package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserDataMatchName(t *testing.T) {
	testcases := map[string]struct {
		userData   UserData
		firstNames string
		lastName   string
		expected   bool
	}{
		"match": {
			userData: UserData{
				FirstNames: "A BEE",
				LastName:   "SEA",
			},
			firstNames: "A Bee",
			lastName:   "Sea",
			expected:   true,
		},
		"match on unordered firstnames": {
			userData: UserData{
				FirstNames: "BEE A",
				LastName:   "SEA",
			},
			firstNames: "A Bee",
			lastName:   "Sea",
			expected:   true,
		},
		"match on accented characters": {
			userData: UserData{
				FirstNames: "A BEE",
				LastName:   "SEA",
			},
			firstNames: "Ã Béë",
			lastName:   "Sêâ",
			expected:   true,
		},
		// TODO if we find out it is worth doing
		// "match on alternate accented characters": {
		//	userData: UserData{
		//		FirstNames: "A BEE",
		//		LastName:   "SEA",
		//	},
		//	firstNames: "Å Béë",
		//	lastName:   "Sêå",
		//	expected:   true,
		// },
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.userData.MatchName(tc.firstNames, tc.lastName))
		})
	}
}
