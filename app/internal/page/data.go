package page

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
)

const (
	PayCookieName              = "pay"
	PayCookiePaymentIdValueKey = "paymentId"
	CostOfLpaPence             = 8200
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

type Lpa struct {
	ID                       string
	You                      Person
	Attorney                 Attorney
	CertificateProvider      CertificateProvider
	WhoFor                   string
	Contact                  []string
	Type                     string
	WantReplacementAttorneys string
	WhenCanTheLpaBeUsed      string
	Restrictions             string
	Tasks                    Tasks
	Checked                  bool
	HappyToShare             bool
	PaymentDetails           PaymentDetails
	CheckedAgain             bool
	ConfirmFreeWill          bool
	SignatureCode            string
	IdentityOptions          IdentityOptions
	YotiUserData             identity.UserData
}

type PaymentDetails struct {
	PaymentReference string
	PaymentId        string
}

type Tasks struct {
	WhenCanTheLpaBeUsed        TaskState
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  TaskState
	ConfirmYourIdentityAndSign TaskState
}

type Person struct {
	FirstNames  string
	LastName    string
	Email       string
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

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Email                   string
	DateOfBirth             time.Time
	Relationship            []string
	RelationshipDescription string
	RelationshipLength      string
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

func readDate(t time.Time) Date {
	return Date{
		Day:   t.Format("2"),
		Month: t.Format("1"),
		Year:  t.Format("2006"),
	}
}

type LpaStore interface {
	Get(context.Context, string) (Lpa, error)
	Put(context.Context, string, Lpa) error
}

type lpaStore struct {
	dataStore DataStore
	randomInt func(int) int
}

func (s *lpaStore) Get(ctx context.Context, sessionID string) (Lpa, error) {
	var lpa Lpa
	if err := s.dataStore.Get(ctx, sessionID, &lpa); err != nil {
		return lpa, err
	}

	if lpa.ID == "" {
		lpa.ID = "10" + strconv.Itoa(s.randomInt(100000))
	}

	return lpa, nil
}

func (s *lpaStore) Put(ctx context.Context, sessionID string, lpa Lpa) error {
	return s.dataStore.Put(ctx, sessionID, lpa)
}
