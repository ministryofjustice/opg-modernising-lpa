package page

import (
	"encoding/json"
	"strings"
	"time"
)

type Donor struct {
	FirstName   string
	LastName    string
	DateOfBirth time.Time
	Address     Address
}

type Address struct {
	Line1      string
	Line2      string
	TownOrCity string
	Postcode   string
}

type AddressClient interface {
	LookupPostcode(string) ([]Address, error)
}

func (a Address) Encode() string {
	x, _ := json.Marshal(a)
	return string(x)
}

func DecodeAddress(s string) *Address {
	var v Address
	json.Unmarshal([]byte(s), &v)
	return &v
}

func (a Address) String() string {
	var parts []string

	if a.Line1 != "" {
		parts = append(parts, a.Line1)
	}
	if a.Line2 != "" {
		parts = append(parts, a.Line2)
	}
	if a.TownOrCity != "" {
		parts = append(parts, a.TownOrCity)
	}
	if a.Postcode != "" {
		parts = append(parts, a.Postcode)
	}

	return strings.Join(parts, ", ")
}
