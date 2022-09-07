package page

import (
	"encoding/json"
	"strings"
	"time"
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

type Lpa struct {
	You                      Person
	Attorney                 Attorney
	WhoFor                   string
	Contact                  []string
	Type                     string
	WantReplacementAttorneys string
	WhenCanTheLpaBeUsed      string
	Restrictions             string
	Tasks                    Tasks
}

type Tasks struct {
	WhenCanTheLpaBeUsed                 TaskState
	Restrictions                        TaskState
	WhoDoYouWantToBeCertificateProvider TaskState
}

type Person struct {
	FirstNames  string
	LastName    string
	OtherNames  string
	DateOfBirth time.Time
	Address     Address
}

type Attorney struct {
	FirstNames  string
	LastName    string
	Email       string
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

type Date struct {
	Day   string
	Month string
	Year  string
}
