package place

import (
	"encoding/json"
	"strings"
)

type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2" relatedhash:"-"`
	Line3      string `json:"line3" relatedhash:"-"`
	TownOrCity string `json:"town" relatedhash:"-"`
	Postcode   string `json:"postcode"`
	Country    string `json:"country" relatedhash:"-"`
}

func (a Address) Encode() string {
	x, _ := json.Marshal(a)
	return string(x)
}

func (a Address) Lines() []string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.Line3 != "" {
		parts = append(parts, a.Line3)
	}
	if a.TownOrCity != "" {
		parts = append(parts, a.TownOrCity)
	}
	if a.Postcode != "" {
		parts = append(parts, a.Postcode)
	}

	return parts
}

func (a Address) String() string {
	return strings.Join(a.Lines(), ", ")
}

// Equal provides a looser definition of equality that we use for various
// checks which need to not be strict.
func (a Address) Equal(b Address) bool {
	return strings.EqualFold(a.Line1, b.Line1) && strings.EqualFold(a.Postcode, b.Postcode)
}
