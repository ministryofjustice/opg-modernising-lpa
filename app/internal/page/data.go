package page

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
)

const (
	AllCanNoLongerAct                = "all"
	CostOfLpaPence                   = 8200
	Jointly                          = "jointly"
	JointlyAndSeverally              = "jointly-and-severally"
	JointlyForSomeSeverallyForOthers = "mixed"
	LpaTypeCombined                  = "both"
	LpaTypeHealthWelfare             = "hw"
	LpaTypePropertyFinance           = "pfa"
	PayCookieName                    = "pay"
	PayCookiePaymentIdValueKey       = "paymentId"
	OneCanNoLongerAct                = "one"
	SomeOtherWay                     = "other"
	UsedWhenCapacityLost             = "when-capacity-lost"
	UsedWhenRegistered               = "when-registered"
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

type Lpa struct {
	ID                                          string
	UpdatedAt                                   time.Time
	You                                         Person
	Attorneys                                   []Attorney
	CertificateProvider                         CertificateProvider
	WhoFor                                      string
	Contact                                     []string
	Type                                        string
	WantReplacementAttorneys                    string
	WhenCanTheLpaBeUsed                         string
	Restrictions                                string
	Tasks                                       Tasks
	Checked                                     bool
	HappyToShare                                bool
	PaymentDetails                              PaymentDetails
	CheckedAgain                                bool
	ConfirmFreeWill                             bool
	SignatureCode                               string
	EnteredSignatureCode                        string
	SignatureEmailID                            string
	IdentityOptions                             IdentityOptions
	YotiUserData                                identity.UserData
	HowAttorneysMakeDecisions                   string
	HowAttorneysMakeDecisionsDetails            string
	ReplacementAttorneys                        []Attorney
	HowReplacementAttorneysMakeDecisions        string
	HowReplacementAttorneysMakeDecisionsDetails string
	HowShouldReplacementAttorneysStepIn         string
	HowShouldReplacementAttorneysStepInDetails  string
	DoYouWantToNotifyPeople                     string
	PeopleToNotify                              []PersonToNotify
	CPWitnessedDonorSign                        bool
	WantToApplyForLpa                           bool
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
	PeopleToNotify             TaskState
}

type Person struct {
	FirstNames  string
	LastName    string
	Email       string
	OtherNames  string
	DateOfBirth time.Time
	Address     place.Address
}

type PersonToNotify struct {
	FirstNames string
	LastName   string
	Email      string
	Address    place.Address
	ID         string
}

type Attorney struct {
	ID          string
	FirstNames  string
	LastName    string
	Email       string
	DateOfBirth time.Time
	Address     place.Address
}

type CertificateProvider struct {
	FirstNames              string
	LastName                string
	Email                   string
	DateOfBirth             time.Time
	Relationship            string
	RelationshipDescription string
	RelationshipLength      string
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
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
	Get(context.Context, string) (*Lpa, error)
	Put(context.Context, string, *Lpa) error
}

type lpaStore struct {
	dataStore DataStore
	randomInt func(int) int
}

func (s *lpaStore) Get(ctx context.Context, sessionID string) (*Lpa, error) {
	var lpa Lpa
	if err := s.dataStore.Get(ctx, sessionID, &lpa); err != nil {
		return &lpa, err
	}

	if lpa.ID == "" {
		lpa.ID = "10" + strconv.Itoa(s.randomInt(100000))
	}

	return &lpa, nil
}

func (s *lpaStore) Put(ctx context.Context, sessionID string, lpa *Lpa) error {
	lpa.UpdatedAt = time.Now()

	return s.dataStore.Put(ctx, sessionID, lpa)
}

func DecodeAddress(s string) *place.Address {
	var v place.Address
	json.Unmarshal([]byte(s), &v)
	return &v
}

func (l *Lpa) GetAttorney(id string) (Attorney, bool) {
	idx := slices.IndexFunc(l.Attorneys, func(a Attorney) bool { return a.ID == id })

	if idx == -1 {
		return Attorney{}, false
	}

	return l.Attorneys[idx], true
}

func (l *Lpa) PutAttorney(attorney Attorney) bool {
	idx := slices.IndexFunc(l.Attorneys, func(a Attorney) bool { return a.ID == attorney.ID })

	if idx == -1 {
		return false
	}

	l.Attorneys[idx] = attorney

	return true
}

func (l *Lpa) DeleteAttorney(attorney Attorney) bool {
	idx := slices.IndexFunc(l.Attorneys, func(a Attorney) bool { return a.ID == attorney.ID })

	if idx == -1 {
		return false
	}

	l.Attorneys = slices.Delete(l.Attorneys, idx, idx+1)

	return true
}

func (l *Lpa) GetReplacementAttorney(id string) (Attorney, bool) {
	idx := slices.IndexFunc(l.ReplacementAttorneys, func(a Attorney) bool { return a.ID == id })

	if idx == -1 {
		return Attorney{}, false
	}

	return l.ReplacementAttorneys[idx], true
}

func (l *Lpa) PutReplacementAttorney(attorney Attorney) bool {
	idx := slices.IndexFunc(l.ReplacementAttorneys, func(a Attorney) bool { return a.ID == attorney.ID })

	if idx == -1 {
		return false
	}

	l.ReplacementAttorneys[idx] = attorney

	return true
}

func (l *Lpa) DeleteReplacementAttorney(attorney Attorney) bool {
	idx := slices.IndexFunc(l.ReplacementAttorneys, func(a Attorney) bool { return a.ID == attorney.ID })

	if idx == -1 {
		return false
	}

	l.ReplacementAttorneys = slices.Delete(l.ReplacementAttorneys, idx, idx+1)

	return true
}

func (l *Lpa) GetPersonToNotify(id string) (PersonToNotify, bool) {
	idx := slices.IndexFunc(l.PeopleToNotify, func(p PersonToNotify) bool { return p.ID == id })

	if idx == -1 {
		return PersonToNotify{}, false
	}

	return l.PeopleToNotify[idx], true
}

func (l *Lpa) PutPersonToNotify(person PersonToNotify) bool {
	idx := slices.IndexFunc(l.PeopleToNotify, func(p PersonToNotify) bool { return p.ID == person.ID })

	if idx == -1 {
		return false
	}

	l.PeopleToNotify[idx] = person

	return true
}

func (l *Lpa) DeletePersonToNotify(personToNotify PersonToNotify) bool {
	idx := slices.IndexFunc(l.PeopleToNotify, func(p PersonToNotify) bool { return p.ID == personToNotify.ID })

	if idx == -1 {
		return false
	}

	l.PeopleToNotify = slices.Delete(l.PeopleToNotify, idx, idx+1)

	return true
}

func (l *Lpa) AttorneysFullNames() string {
	var names []string

	for _, a := range l.Attorneys {
		names = append(names, fmt.Sprintf("%s %s", a.FirstNames, a.LastName))
	}

	return concatSentence(names)
}

func (l *Lpa) AttorneysFirstNames() string {
	var names []string

	for _, a := range l.Attorneys {
		names = append(names, a.FirstNames)
	}

	return concatSentence(names)
}

func (l *Lpa) ReplacementAttorneysFullNames() string {
	var names []string

	for _, a := range l.ReplacementAttorneys {
		names = append(names, fmt.Sprintf("%s %s", a.FirstNames, a.LastName))
	}

	return concatSentence(names)
}

func (l *Lpa) ReplacementAttorneysFirstNames() string {
	var names []string

	for _, a := range l.ReplacementAttorneys {
		names = append(names, a.FirstNames)
	}

	return concatSentence(names)
}

func concatSentence(list []string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		last := len(list) - 1
		return fmt.Sprintf("%s and %s", strings.Join(list[:last], ", "), list[last])
	}
}

func (l *Lpa) ReplacementAttorneysTaskComplete() bool {
	//"replacement attorneys not required"
	if l.WantReplacementAttorneys == "no" && len(l.ReplacementAttorneys) == 0 {
		return true
	}

	if !l.ReplacementAttorneysValid() {
		return false
	}

	if l.WantReplacementAttorneys == "yes" {
		if len(l.Attorneys) == 1 {
			//"single attorney and single replacement attorney"
			if len(l.ReplacementAttorneys) == 1 {
				return true
			}

			//"single attorney and multiple replacement attorney acting jointly"
			//"single attorney and multiple replacement attorney acting jointly and severally"
			//"single attorney and multiple replacement attorneys acting mixed with details"
			if len(l.ReplacementAttorneys) > 1 {
				return l.ReplacementAttorneysActJointlyOrJointlyAndSeverally() || l.ReplacementAttorneysActJointlyForSomeSeverallyForOthersWithDetails()
			}
		}

		if len(l.Attorneys) > 1 {
			//"multiple attorneys acting jointly and severally and single replacement attorney steps in when there are no attorneys left to act"
			//"multiple attorneys acting jointly and severally and single replacement attorney steps in when one attorney can no longer act"
			//"multiple attorneys acting jointly and severally and single replacement attorney steps in in some other way with details"
			//"multiple attorneys acting jointly and severally and multiple replacement attorneys acting jointly steps in when there are no attorneys left to act"
			//"multiple attorneys acting jointly and severally and multiple replacement attorney acting jointly and severally steps in when there are no attorneys left to act"
			//"multiple attorneys acting jointly and severally and multiple replacement attorney acting mixed with details steps in when there are no attorneys left to act"
			//"multiple attorneys acting jointly and severally and multiple replacement attorneys steps in when one attorney cannot act"
			if l.HowAttorneysMakeDecisions == JointlyAndSeverally &&
				len(l.ReplacementAttorneys) > 0 {
				return l.ReplacementAttorneysStepInWhenOneOrAllAttorneysCannotAct() || l.ReplacementAttorneysStepInSomeOtherWayWithDetails()
			}

			//"multiple attorneys acting mixed with details and single replacement attorney with blank how to step in"
			//"multiple attorneys acting mixed with details and multiple replacement attorney with blank how to step in"
			if l.AttorneysActJointlyForSomeSeverallyForOthersWithDetails() &&
				len(l.ReplacementAttorneys) > 0 &&
				l.HowShouldReplacementAttorneysStepIn == "" {
				return true
			}

			if l.HowAttorneysMakeDecisions == Jointly {
				//"multiple attorneys acting jointly and multiple replacement attorneys acting jointly and blank how to step in"
				//"multiple attorneys acting jointly and multiple replacement attorneys acting jointly and severally and blank how to step in"
				//"multiple attorneys acting jointly and multiple replacement attorneys acting mixed with details and blank how to step in"
				if len(l.ReplacementAttorneys) > 1 &&
					(l.ReplacementAttorneysActJointlyOrJointlyAndSeverally() || l.ReplacementAttorneysActJointlyForSomeSeverallyForOthersWithDetails()) &&
					l.HowShouldReplacementAttorneysStepIn == "" {
					return true
				}

				//"multiple attorneys acting jointly and single replacement attorneys and blank how to step in"
				if len(l.ReplacementAttorneys) == 1 &&
					l.HowShouldReplacementAttorneysStepIn == "" {
					return true
				}

			}
		}
	}

	return false
}

func (l *Lpa) AttorneysTaskComplete() bool {
	if len(l.Attorneys) == 0 {
		return false
	}

	if !l.AttorneysValid() {
		return false
	}

	if l.AttorneysActJointlyOrJointlyAndSeverally() ||
		l.AttorneysActJointlyForSomeSeverallyForOthersWithDetails() ||
		len(l.Attorneys) == 1 {
		return true
	}

	return false
}

func (l *Lpa) AttorneysValid() bool {
	for _, a := range l.Attorneys {
		if a.Address.Line1 == "" || a.FirstNames == "" || a.LastName == "" || a.DateOfBirth.IsZero() {
			return false
		}
	}

	return true
}

func (l *Lpa) ReplacementAttorneysValid() bool {
	for _, a := range l.ReplacementAttorneys {
		if a.Address.Line1 == "" || a.FirstNames == "" || a.LastName == "" || a.DateOfBirth.IsZero() {
			return false
		}
	}

	return true
}

func (l *Lpa) PeopleToNotifyValid() bool {
	for _, a := range l.PeopleToNotify {
		if a.Address.Line1 == "" || a.FirstNames == "" || a.LastName == "" {
			return false
		}
	}

	return true
}

func (l *Lpa) AttorneysActJointlyOrJointlyAndSeverally() bool {
	return slices.Contains([]string{Jointly, JointlyAndSeverally}, l.HowAttorneysMakeDecisions)
}

func (l *Lpa) AttorneysActJointlyForSomeSeverallyForOthersWithDetails() bool {
	return l.HowAttorneysMakeDecisions == JointlyForSomeSeverallyForOthers && l.HowAttorneysMakeDecisionsDetails != ""
}

func (l *Lpa) ReplacementAttorneysActJointlyOrJointlyAndSeverally() bool {
	return slices.Contains([]string{Jointly, JointlyAndSeverally}, l.HowReplacementAttorneysMakeDecisions)
}

func (l *Lpa) ReplacementAttorneysActJointlyForSomeSeverallyForOthersWithDetails() bool {
	return l.HowReplacementAttorneysMakeDecisions == JointlyForSomeSeverallyForOthers &&
		l.HowReplacementAttorneysMakeDecisionsDetails != ""
}

func (l *Lpa) ReplacementAttorneysStepInWhenOneOrAllAttorneysCannotAct() bool {
	return slices.Contains([]string{OneCanNoLongerAct, AllCanNoLongerAct}, l.HowShouldReplacementAttorneysStepIn)
}

func (l *Lpa) ReplacementAttorneysStepInSomeOtherWayWithDetails() bool {
	return l.HowShouldReplacementAttorneysStepIn == SomeOtherWay && l.HowShouldReplacementAttorneysStepInDetails != ""
}

func (l *Lpa) DonorFullName() string {
	return fmt.Sprintf("%s %s", l.You.FirstNames, l.You.LastName)
}

func (l *Lpa) CertificateProviderFullName() string {
	return fmt.Sprintf("%s %s", l.CertificateProvider.FirstNames, l.CertificateProvider.LastName)
}

func (l *Lpa) LpaLegalTerm() string {
	switch l.Type {
	case LpaTypePropertyFinance:
		return "finance and affairs"
	case LpaTypeHealthWelfare:
		return "personal welfare"
	case LpaTypeCombined:
		return "combined"
	}
	return ""
}
