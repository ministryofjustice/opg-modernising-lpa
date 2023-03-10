package page

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
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
	ID                                          string
	UpdatedAt                                   time.Time
	Donor                                       actor.Donor
	Attorneys                                   actor.Attorneys
	CertificateProvider                         actor.CertificateProvider
	WhoFor                                      string
	Type                                        string
	WantReplacementAttorneys                    string
	WhenCanTheLpaBeUsed                         string
	Restrictions                                string
	Tasks                                       Tasks
	Checked                                     bool
	HappyToShare                                bool
	PaymentDetails                              PaymentDetails
	DonorIdentityOption                         identity.Option
	DonorIdentityUserData                       identity.UserData
	HowAttorneysMakeDecisions                   string
	HowAttorneysMakeDecisionsDetails            string
	ReplacementAttorneys                        actor.Attorneys
	HowReplacementAttorneysMakeDecisions        string
	HowReplacementAttorneysMakeDecisionsDetails string
	HowShouldReplacementAttorneysStepIn         string
	HowShouldReplacementAttorneysStepInDetails  string
	DoYouWantToNotifyPeople                     string
	PeopleToNotify                              actor.PeopleToNotify
	WitnessCodes                                WitnessCodes
	WantToApplyForLpa                           bool
	WantToSignLpa                               bool
	Submitted                                   time.Time
	CPWitnessCodeValidated                      bool
	WitnessCodeLimiter                          *Limiter

	CertificateProviderIdentityOption   identity.Option
	CertificateProviderIdentityUserData identity.UserData
	CertificateProviderProvidedDetails  actor.CertificateProvider
	Certificate                         Certificate
}

type PaymentDetails struct {
	PaymentReference string
	PaymentId        string
}

type Tasks struct {
	YourDetails                TaskState
	ChooseAttorneys            TaskState
	ChooseReplacementAttorneys TaskState
	WhenCanTheLpaBeUsed        TaskState
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  TaskState
	ConfirmYourIdentityAndSign TaskState
	PeopleToNotify             TaskState
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

func SessionDataFromContext(ctx context.Context) *SessionData {
	data, _ := ctx.Value((*SessionData)(nil)).(*SessionData)

	return data
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

func (l *Lpa) CertificateProviderIdentityConfirmed() bool {
	return l.CertificateProviderIdentityUserData.OK && l.CertificateProviderIdentityUserData.Provider != identity.UnknownOption &&
		l.CertificateProviderIdentityUserData.MatchName(l.CertificateProvider.FirstNames, l.CertificateProvider.LastName) &&
		l.CertificateProviderIdentityUserData.DateOfBirth.Equals(l.CertificateProvider.DateOfBirth)
}

func (l *Lpa) TypeLegalTermTransKey() string {
	switch l.Type {
	case LpaTypePropertyFinance:
		return "pfaLegalTerm"
	case LpaTypeHealthWelfare:
		return "hwLegalTerm"
	case LpaTypeCombined:
		return "combinedLegalTerm"
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
	case Paths.WhenCanTheLpaBeUsed, Paths.Restrictions, Paths.WhoDoYouWantToBeCertificateProviderGuidance, Paths.DoYouWantToNotifyPeople:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed()
	case Paths.CheckYourLpa:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed() &&
			l.Tasks.ChooseReplacementAttorneys.Completed() &&
			l.Tasks.WhenCanTheLpaBeUsed.Completed() &&
			l.Tasks.Restrictions.Completed() &&
			l.Tasks.CertificateProvider.Completed() &&
			l.Tasks.PeopleToNotify.Completed()
	case Paths.AboutPayment:
		return l.Tasks.YourDetails.Completed() &&
			l.Tasks.ChooseAttorneys.Completed() &&
			l.Tasks.ChooseReplacementAttorneys.Completed() &&
			l.Tasks.WhenCanTheLpaBeUsed.Completed() &&
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

func (l *Lpa) Progress() Progress {
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

	if !l.Certificate.Agreed.IsZero() {
		p.CertificateProviderDeclared = TaskCompleted
		p.AttorneysDeclared = TaskInProgress
	}

	return p
}

type Certificate struct {
	AgreeToStatement bool
	Agreed           time.Time
}
