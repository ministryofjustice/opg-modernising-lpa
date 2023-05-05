package page

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

const (
	AllCanNoLongerAct          = "all"
	CostOfLpaPence             = 8200
	LpaTypeHealthWelfare       = "hw"
	LpaTypePropertyFinance     = "pfa"
	PayCookieName              = "pay"
	PayCookiePaymentIdValueKey = "paymentId"
	OneCanNoLongerAct          = "one"
	SomeOtherWay               = "other"
	UsedWhenCapacityLost       = "when-capacity-lost"
	UsedWhenRegistered         = "when-registered"
	OptionA                    = "option-a"
	OptionB                    = "option-b"
)

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

func (t TaskState) InProgress() bool { return t == TaskInProgress }
func (t TaskState) Completed() bool  { return t == TaskCompleted }

func (t TaskState) String() string {
	switch t {
	case TaskNotStarted:
		return "notStarted"
	case TaskInProgress:
		return "inProgress"
	case TaskCompleted:
		return "completed"
	}
	return ""
}

type Lpa struct {
	ID                                         string
	UpdatedAt                                  time.Time
	Donor                                      actor.Donor
	Attorneys                                  actor.Attorneys
	AttorneyDecisions                          actor.AttorneyDecisions
	CertificateProviderDetails                 CertificateProviderDetails
	WhoFor                                     string
	Type                                       string
	WantReplacementAttorneys                   string
	WhenCanTheLpaBeUsed                        string
	LifeSustainingTreatmentOption              string
	Restrictions                               string
	Tasks                                      Tasks
	Checked                                    bool
	HappyToShare                               bool
	PaymentDetails                             PaymentDetails
	DonorIdentityOption                        identity.Option
	DonorIdentityUserData                      identity.UserData
	ReplacementAttorneys                       actor.Attorneys
	ReplacementAttorneyDecisions               actor.AttorneyDecisions
	HowShouldReplacementAttorneysStepIn        string
	HowShouldReplacementAttorneysStepInDetails string
	DoYouWantToNotifyPeople                    string
	PeopleToNotify                             actor.PeopleToNotify
	WitnessCodes                               WitnessCodes
	WantToApplyForLpa                          bool
	WantToSignLpa                              bool
	Submitted                                  time.Time
	CPWitnessCodeValidated                     bool
	WitnessCodeLimiter                         *Limiter

	AttorneyProvidedDetails            map[string]actor.AttorneyProvidedDetails
	ReplacementAttorneyProvidedDetails map[string]actor.AttorneyProvidedDetails

	AttorneyTasks            map[string]AttorneyTasks
	ReplacementAttorneyTasks map[string]AttorneyTasks
}

type PaymentDetails struct {
	PaymentReference string
	PaymentId        string
}

type CertificateProviderDetails struct {
	FirstNames              string
	LastName                string
	Address                 place.Address
	Mobile                  string
	Email                   string
	CarryOutBy              string
	DateOfBirth             date.Date
	Relationship            string
	RelationshipDescription string
	RelationshipLength      string
}

func (c CertificateProviderDetails) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}

type Tasks struct {
	YourDetails                TaskState
	ChooseAttorneys            TaskState
	ChooseReplacementAttorneys TaskState
	WhenCanTheLpaBeUsed        TaskState // pfa only
	LifeSustainingTreatment    TaskState // hw only
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  TaskState
	ConfirmYourIdentityAndSign TaskState
	PeopleToNotify             TaskState
}

type AttorneyTasks struct {
	ConfirmYourDetails TaskState
	ReadTheLpa         TaskState
	SignTheLpa         TaskState
}

type Progress struct {
	LpaSigned                   TaskState
	CertificateProviderDeclared TaskState
	AttorneysDeclared           TaskState
	LpaSubmitted                TaskState
	StatutoryWaitingPeriod      TaskState
	LpaRegistered               TaskState
}

type SessionData struct {
	SessionID string
	LpaID     string
}

type SessionMissingError struct{}

func (s SessionMissingError) Error() string {
	return "Session data not set"
}

func SessionDataFromContext(ctx context.Context) (*SessionData, error) {
	data, ok := ctx.Value((*SessionData)(nil)).(*SessionData)

	if !ok {
		return nil, SessionMissingError{}
	}

	return data, nil
}

func ContextWithSessionData(ctx context.Context, data *SessionData) context.Context {
	return context.WithValue(ctx, (*SessionData)(nil), data)
}

func DecodeAddress(s string) *place.Address {
	var v place.Address
	json.Unmarshal([]byte(s), &v)
	return &v
}

func (l *Lpa) DonorIdentityConfirmed() bool {
	return l.DonorIdentityUserData.OK && l.DonorIdentityUserData.Provider != identity.UnknownOption &&
		l.DonorIdentityUserData.MatchName(l.Donor.FirstNames, l.Donor.LastName) &&
		l.DonorIdentityUserData.DateOfBirth.Equals(l.Donor.DateOfBirth)
}

func (l *Lpa) TypeLegalTermTransKey() string {
	switch l.Type {
	case LpaTypePropertyFinance:
		return "pfaLegalTerm"
	case LpaTypeHealthWelfare:
		return "hwLegalTerm"
	}
	return ""
}

func (l *Lpa) AttorneysAndCpSigningDeadline() time.Time {
	return l.Submitted.Add((24 * time.Hour) * 28)
}

func (l *Lpa) CanGoTo(url string) bool {
	path, _, _ := strings.Cut(url, "?")

	switch path {
	case Paths.ReadYourLpa, Paths.SignYourLpa, Paths.WitnessingYourSignature, Paths.WitnessingAsCertificateProvider, Paths.YouHaveSubmittedYourLpa:
		return l.DonorIdentityConfirmed()
	case Paths.WhenCanTheLpaBeUsed, Paths.LifeSustainingTreatment, Paths.Restrictions, Paths.WhoDoYouWantToBeCertificateProviderGuidance, Paths.DoYouWantToNotifyPeople:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed()
	case Paths.CheckYourLpa:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed() &&
			l.Tasks.ChooseReplacementAttorneys.Completed() &&
			(l.Type == LpaTypeHealthWelfare && l.Tasks.LifeSustainingTreatment.Completed() || l.Tasks.WhenCanTheLpaBeUsed.Completed()) &&
			l.Tasks.Restrictions.Completed() &&
			l.Tasks.CertificateProvider.Completed() &&
			l.Tasks.PeopleToNotify.Completed()
	case Paths.AboutPayment:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed() &&
			l.Tasks.ChooseReplacementAttorneys.Completed() &&
			(l.Type == LpaTypeHealthWelfare && l.Tasks.LifeSustainingTreatment.Completed() || l.Tasks.WhenCanTheLpaBeUsed.Completed()) &&
			l.Tasks.Restrictions.Completed() &&
			l.Tasks.CertificateProvider.Completed() &&
			l.Tasks.PeopleToNotify.Completed() &&
			l.Tasks.CheckYourLpa.Completed()
	case Paths.SelectYourIdentityOptions, Paths.HowToConfirmYourIdentityAndSign:
		return l.Tasks.PayForLpa.Completed()
	case "":
		return false
	default:
		return true
	}
}

func (l *Lpa) Progress(certificateProvider *actor.CertificateProvider) Progress {
	p := Progress{
		LpaSigned:                   TaskInProgress,
		CertificateProviderDeclared: TaskNotStarted,
		AttorneysDeclared:           TaskNotStarted,
		LpaSubmitted:                TaskNotStarted,
		StatutoryWaitingPeriod:      TaskNotStarted,
		LpaRegistered:               TaskNotStarted,
	}

	if !l.Submitted.IsZero() {
		p.LpaSigned = TaskCompleted
		p.CertificateProviderDeclared = TaskInProgress
	}

	if !certificateProvider.Certificate.Agreed.IsZero() {
		p.CertificateProviderDeclared = TaskCompleted
		p.AttorneysDeclared = TaskInProgress
	}

	return p
}

type AddressDetail struct {
	Name    string
	Role    actor.Type
	Address place.Address
	ID      string
}

func (l *Lpa) ActorAddresses() []AddressDetail {
	var ads []AddressDetail

	if l.Donor.Address.String() != "" {
		ads = append(ads, AddressDetail{
			Name:    l.Donor.FullName(),
			Role:    actor.TypeDonor,
			Address: l.Donor.Address,
		})
	}

	if l.CertificateProviderDetails.Address.String() != "" {
		ads = append(ads, AddressDetail{
			Name:    l.CertificateProviderDetails.FullName(),
			Role:    actor.TypeCertificateProvider,
			Address: l.CertificateProviderDetails.Address,
		})
	}

	for _, attorney := range l.Attorneys {
		if attorney.Address.String() != "" {
			ads = append(ads, AddressDetail{
				Name:    fmt.Sprintf("%s %s", attorney.FirstNames, attorney.LastName),
				Role:    actor.TypeAttorney,
				Address: attorney.Address,
				ID:      attorney.ID,
			})
		}
	}

	for _, replacementAttorney := range l.ReplacementAttorneys {
		if replacementAttorney.Address.String() != "" {
			ads = append(ads, AddressDetail{
				Name:    fmt.Sprintf("%s %s", replacementAttorney.FirstNames, replacementAttorney.LastName),
				Role:    actor.TypeReplacementAttorney,
				Address: replacementAttorney.Address,
				ID:      replacementAttorney.ID,
			})
		}
	}

	return ads
}

func ChooseAttorneysState(attorneys actor.Attorneys, decisions actor.AttorneyDecisions) TaskState {
	if len(attorneys) == 0 {
		return TaskNotStarted
	}

	for _, a := range attorneys {
		if a.FirstNames == "" || (a.Address.Line1 == "" && a.Email == "") {
			return TaskInProgress
		}
	}

	if len(attorneys) > 1 && !decisions.IsComplete(len(attorneys)) {
		return TaskInProgress
	}

	return TaskCompleted
}

func ChooseReplacementAttorneysState(lpa *Lpa) TaskState {
	if lpa.WantReplacementAttorneys == "no" {
		return TaskCompleted
	}

	if len(lpa.ReplacementAttorneys) == 0 {
		if lpa.WantReplacementAttorneys == "" {
			return TaskNotStarted
		}

		return TaskInProgress
	}

	for _, a := range lpa.ReplacementAttorneys {
		if a.FirstNames == "" || (a.Address.Line1 == "" && a.Email == "") {
			return TaskInProgress
		}
	}

	if len(lpa.ReplacementAttorneys) > 1 &&
		lpa.HowShouldReplacementAttorneysStepIn != OneCanNoLongerAct &&
		!lpa.ReplacementAttorneyDecisions.IsComplete(len(lpa.ReplacementAttorneys)) {
		return TaskInProgress
	}

	if lpa.AttorneyDecisions.How == actor.Jointly &&
		len(lpa.ReplacementAttorneys) > 1 &&
		!lpa.ReplacementAttorneyDecisions.IsComplete(len(lpa.ReplacementAttorneys)) {
		return TaskInProgress
	}

	if lpa.AttorneyDecisions.How == actor.JointlyAndSeverally {
		if lpa.HowShouldReplacementAttorneysStepIn == "" {
			return TaskInProgress
		}

		if len(lpa.ReplacementAttorneys) > 1 &&
			lpa.HowShouldReplacementAttorneysStepIn == AllCanNoLongerAct &&
			!lpa.ReplacementAttorneyDecisions.IsComplete(len(lpa.ReplacementAttorneys)) {
			return TaskInProgress
		}
	}

	return TaskCompleted
}

func (l *Lpa) IsHealthAndWelfareLpa() bool {
	return l.Type == LpaTypeHealthWelfare
}
