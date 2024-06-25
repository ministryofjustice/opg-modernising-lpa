package identity

import (
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

// https://www.icao.int/publications/Documents/9303_p3_cons_en.pdf
var charmap = map[rune][]rune{
	'\u00C0': {'A'},
	'\u00C1': {'A'},
	'\u00C2': {'A'},
	'\u00C3': {'A'},
	'\u00C4': {'A', 'E'}, // or "A"
	'\u00C5': {'A', 'A'}, // or "A"
	'\u00C6': {'A', 'E'},
	'\u00C7': {'C'},
	'\u00C8': {'E'},
	'\u00C9': {'E'},
	'\u00CA': {'E'},
	'\u00CB': {'E'},
	'\u00CC': {'I'},
	'\u00CD': {'I'},
	'\u00CE': {'I'},
	'\u00CF': {'I'},
	'\u00D0': {'D'},
	'\u00D1': {'N'}, // or "NXX"
	'\u00D2': {'O'},
	'\u00D3': {'O'},
	'\u00D4': {'O'},
	'\u00D5': {'O'},
	'\u00D6': {'O', 'E'}, // or "O"
	'\u00D8': {'O', 'E'},
	'\u00D9': {'U'},
	'\u00DA': {'U'},
	'\u00DB': {'U'},
	'\u00DC': {'U', 'E'}, // or "UXX" or "U"
	'\u00DD': {'Y'},
	'\u00DE': {'T', 'H'},
	// end page 1
	'\u0100': {'A'},
	'\u0102': {'A'},
	'\u0104': {'A'},
	'\u0106': {'C'},
	'\u0108': {'C'},
	'\u010A': {'C'},
	'\u010C': {'C'},
	'\u010E': {'D'},
	'\u0110': {'D'},
	'\u0112': {'E'},
	'\u0114': {'E'},
	'\u0116': {'E'},
	'\u0118': {'E'},
	'\u011A': {'E'},
	'\u011C': {'G'},
	'\u011E': {'G'},
	'\u0120': {'G'},
	'\u0122': {'G'},
	'\u0124': {'H'},
	'\u0126': {'H'},
	'\u0128': {'I'},
	'\u012A': {'I'},
	'\u012C': {'I'},
	'\u012E': {'I'},
	'\u0130': {'I'},
	'\u0131': {'I'},
	'\u0132': {'I', 'J'},
	'\u0134': {'J'},
	'\u0136': {'K'},
	'\u0139': {'L'},
	'\u013B': {'L'},
	'\u013D': {'L'},
	'\u013F': {'L'},
	'\u0141': {'L'},
	'\u0143': {'N'},
	'\u0145': {'N'},
	'\u0147': {'N'},
	// end page 2
	'\u014A': {'N'},
	'\u014C': {'O'},
	'\u014E': {'O'},
	'\u0150': {'O'},
	'\u0152': {'O', 'E'},
	'\u0154': {'R'},
	'\u0156': {'R'},
	'\u0158': {'R'},
	'\u015A': {'S'},
	'\u015C': {'S'},
	'\u015E': {'S'},
	'\u0160': {'S'},
	'\u0162': {'T'},
	'\u0164': {'T'},
	'\u0166': {'T'},
	'\u0168': {'U'},
	'\u016A': {'U'},
	'\u016C': {'U'},
	'\u016E': {'U'},
	'\u0170': {'U'},
	'\u0172': {'U'},
	'\u0174': {'W'},
	'\u0176': {'Y'},
	'\u0178': {'Y'},
	'\u0179': {'Z'},
	'\u017B': {'Z'},
	'\u017D': {'Z'},
	'\u1E9E': {'S', 'S'},
	// end page 3
}

type UserData struct {
	Status      IdentityStatus
	FirstNames  string
	LastName    string
	DateOfBirth date.Date
	RetrievedAt time.Time
}

func (u UserData) MatchName(firstNames, lastName string) bool {
	expectedFirstNames := strings.ToUpper(u.FirstNames)
	expectedLastName := strings.ToUpper(u.LastName)

	firstNames = strings.ToUpper(firstNames)
	lastName = strings.ToUpper(lastName)

	if expectedFirstNames == firstNames && expectedLastName == lastName {
		return true
	}

	expectedFirstNameParts := strings.Fields(expectedFirstNames)
	firstNameParts := strings.Fields(transliterate(firstNames))

	slices.Sort(expectedFirstNameParts)
	slices.Sort(firstNameParts)

	return slices.Equal(expectedFirstNameParts, firstNameParts) && expectedLastName == transliterate(lastName)
}

func transliterate(s string) string {
	var runes []rune
	for _, r := range s {
		if c, ok := charmap[r]; ok {
			runes = append(runes, c...)
		} else {
			runes = append(runes, r)
		}
	}

	return string(runes)
}
