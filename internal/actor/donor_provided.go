package actor

import (
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/mitchellh/hashstructure/v2"
)

//go:generate enumerator -type LpaType -linecomment -trimprefix -empty
type LpaType uint8

const (
	LpaTypeHealthWelfare   LpaType = iota + 1 // hw
	LpaTypePropertyFinance                    // pfa
)

func (e LpaType) LegalTermTransKey() string {
	switch e {
	case LpaTypePropertyFinance:
		return "pfaLegalTerm"
	case LpaTypeHealthWelfare:
		return "hwLegalTerm"
	}
	return ""
}

//go:generate enumerator -type CanBeUsedWhen -linecomment -trimprefix -empty
type CanBeUsedWhen uint8

const (
	CanBeUsedWhenCapacityLost CanBeUsedWhen = iota + 1 // when-capacity-lost
	CanBeUsedWhenHasCapacity                           // when-has-capacity
)

//go:generate enumerator -type LifeSustainingTreatment -linecomment -trimprefix -empty
type LifeSustainingTreatment uint8

const (
	LifeSustainingTreatmentOptionA LifeSustainingTreatment = iota + 1 // option-a
	LifeSustainingTreatmentOptionB                                    // option-b
)

//go:generate enumerator -type ReplacementAttorneysStepIn -linecomment -trimprefix -empty
type ReplacementAttorneysStepIn uint8

const (
	ReplacementAttorneysStepInWhenAllCanNoLongerAct ReplacementAttorneysStepIn = iota + 1 // all
	ReplacementAttorneysStepInWhenOneCanNoLongerAct                                       // one
	ReplacementAttorneysStepInAnotherWay                                                  // other
)

type Payment struct {
	// Reference generated for the payment
	PaymentReference string
	// ID returned from GOV.UK Pay
	PaymentId string
	// Amount is the amount paid in pence
	Amount int
}

type DonorTasks struct {
	YourDetails                TaskState
	ChooseAttorneys            TaskState
	ChooseReplacementAttorneys TaskState
	WhenCanTheLpaBeUsed        TaskState // pfa only
	LifeSustainingTreatment    TaskState // hw only
	Restrictions               TaskState
	CertificateProvider        TaskState
	CheckYourLpa               TaskState
	PayForLpa                  PaymentTask
	ConfirmYourIdentityAndSign TaskState
	ChooseYourSignatory        TaskState // if .Donor.CanSign.IsNo only
	PeopleToNotify             TaskState
}

// DonorProvidedDetails contains all the data related to the LPA application
type DonorProvidedDetails struct {
	PK, SK string
	// Hash is used to determine whether the Lpa has been changed since last read
	Hash uint64 `hash:"-"`
	// Identifies the LPA being drafted
	ID string
	// A unique identifier created after sending basic LPA details to the UID service
	UID string `dynamodbav:",omitempty"`
	// CreatedAt is when the LPA was created
	CreatedAt time.Time
	// UpdatedAt is when the LPA was last updated
	UpdatedAt time.Time `hash:"-"`
	// The donor the LPA relates to
	Donor Donor
	// Attorneys named in the LPA
	Attorneys Attorneys
	// Information on how the applicant wishes their attorneys to act
	AttorneyDecisions AttorneyDecisions
	// The certificate provider named in the LPA
	CertificateProvider CertificateProvider
	// Type of LPA being drafted
	Type LpaType
	// Whether the applicant wants to add replacement attorneys
	WantReplacementAttorneys form.YesNo
	// When the LPA can be used
	WhenCanTheLpaBeUsed CanBeUsedWhen
	// Preferences on life sustaining treatment (applicable to personal welfare LPAs only)
	LifeSustainingTreatmentOption LifeSustainingTreatment
	// Restrictions on attorneys actions
	Restrictions string
	// Used to show the task list
	Tasks DonorTasks
	// PaymentDetails are records of payments made for the LPA via GOV.UK Pay
	PaymentDetails []Payment
	// Information returned by the identity service related to the applicant
	DonorIdentityUserData identity.UserData
	// Replacement attorneys named in the LPA
	ReplacementAttorneys Attorneys
	// Information on how the applicant wishes their replacement attorneys to act
	ReplacementAttorneyDecisions AttorneyDecisions
	// How to bring in replacement attorneys, if set
	HowShouldReplacementAttorneysStepIn ReplacementAttorneysStepIn
	// Details on how replacement attorneys must step in if HowShouldReplacementAttorneysStepIn is set to "other"
	HowShouldReplacementAttorneysStepInDetails string
	// Whether the applicant wants to notify people about the application
	DoYouWantToNotifyPeople form.YesNo
	// People to notify about the application
	PeopleToNotify PeopleToNotify
	// The AuthorisedSignatory signs on the donor's behalf if they are unable to sign
	AuthorisedSignatory AuthorisedSignatory
	// The IndependentWitness acts as an additional witness when the LPA is signed
	IndependentWitness IndependentWitness
	// Confirmation that the applicant wants to apply to register the LPA
	WantToApplyForLpa bool
	// Confirmation that the applicant wants to sign the LPA
	WantToSignLpa bool
	// CheckedAt is when the donor checked their LPA
	CheckedAt time.Time
	// CheckedHash is the Hash value of the LPA when last checked
	CheckedHash uint64 `hash:"-"`
	// SignedAt is when the donor submitted their signature
	SignedAt time.Time
	// SubmittedAt is when the Lpa was sent to the OPG
	SubmittedAt time.Time
	// RegisteredAt is when the Lpa was registered by the OPG
	RegisteredAt time.Time
	// WithdrawnAt is when the Lpa was withdrawn by the donor
	WithdrawnAt time.Time
	// Version is the number of times the LPA has been updated (auto-incremented on PUT)
	Version int `hash:"-"`

	// Codes used for the certificate provider to witness signing
	CertificateProviderCodes WitnessCodes
	// When the signing was witnessed by the certificate provider
	WitnessedByCertificateProviderAt time.Time
	// Codes used for the independent witness to witness signing
	IndependentWitnessCodes WitnessCodes
	// When the signing was witnessed by the independent witness
	WitnessedByIndependentWitnessAt time.Time
	// Used to rate limit witness code attempts
	WitnessCodeLimiter *Limiter

	// FeeType is the type of fee the user is applying for
	FeeType pay.FeeType
	// EvidenceDelivery is the method by which the user wants to send evidence
	EvidenceDelivery pay.EvidenceDelivery
	// PreviousApplicationNumber if the application is related to an existing application
	PreviousApplicationNumber string
	// PreviousFee is the fee previously paid for an LPA
	PreviousFee pay.PreviousFee

	HasSentUidRequestedEvent              bool `hash:"-"`
	HasSentApplicationUpdatedEvent        bool `hash:"-"`
	HasSentPreviousApplicationLinkedEvent bool `hash:"-"`
}

func (l *DonorProvidedDetails) GenerateHash() (uint64, error) {
	return hashstructure.Hash(l, hashstructure.FormatV2, nil)
}

func (l *DonorProvidedDetails) DonorIdentityConfirmed() bool {
	return l.DonorIdentityUserData.OK &&
		l.DonorIdentityUserData.MatchName(l.Donor.FirstNames, l.Donor.LastName) &&
		l.DonorIdentityUserData.DateOfBirth.Equals(l.Donor.DateOfBirth)
}

func (l *DonorProvidedDetails) AttorneysAndCpSigningDeadline() time.Time {
	return l.SignedAt.Add((24 * time.Hour) * 28)
}

type Progress struct {
	DonorSigned               TaskState
	CertificateProviderSigned TaskState
	AttorneysSigned           TaskState
	LpaSubmitted              TaskState
	StatutoryWaitingPeriod    TaskState
	LpaRegistered             TaskState
}

func (l *DonorProvidedDetails) Progress(certificateProvider *CertificateProviderProvidedDetails, attorneys []*AttorneyProvidedDetails) Progress {
	p := Progress{
		DonorSigned:               TaskInProgress,
		CertificateProviderSigned: TaskNotStarted,
		AttorneysSigned:           TaskNotStarted,
		LpaSubmitted:              TaskNotStarted,
		StatutoryWaitingPeriod:    TaskNotStarted,
		LpaRegistered:             TaskNotStarted,
	}

	if l.SignedAt.IsZero() {
		return p
	}

	p.DonorSigned = TaskCompleted
	p.CertificateProviderSigned = TaskInProgress

	if !certificateProvider.Signed(l.SignedAt) {
		return p
	}

	p.CertificateProviderSigned = TaskCompleted
	p.AttorneysSigned = TaskInProgress

	if !l.AllAttorneysSigned(attorneys) {
		return p
	}

	p.AttorneysSigned = TaskCompleted
	p.LpaSubmitted = TaskInProgress

	if l.SubmittedAt.IsZero() {
		return p
	}

	p.LpaSubmitted = TaskCompleted
	p.StatutoryWaitingPeriod = TaskInProgress

	if l.RegisteredAt.IsZero() {
		return p
	}

	p.StatutoryWaitingPeriod = TaskCompleted
	p.LpaRegistered = TaskCompleted

	return p
}

func (l *DonorProvidedDetails) AllAttorneysSigned(attorneys []*AttorneyProvidedDetails) bool {
	if l == nil || l.SignedAt.IsZero() || l.Attorneys.Len() == 0 {
		return false
	}

	var (
		attorneysSigned                   = map[string]struct{}{}
		replacementAttorneysSigned        = map[string]struct{}{}
		trustCorporationSigned            = false
		replacementTrustCorporationSigned = false
	)

	for _, a := range attorneys {
		if !a.Signed(l.SignedAt) {
			continue
		}

		if a.IsReplacement && a.IsTrustCorporation {
			replacementTrustCorporationSigned = true
		} else if a.IsReplacement {
			replacementAttorneysSigned[a.ID] = struct{}{}
		} else if a.IsTrustCorporation {
			trustCorporationSigned = true
		} else {
			attorneysSigned[a.ID] = struct{}{}
		}
	}

	if l.ReplacementAttorneys.TrustCorporation.Name != "" && !replacementTrustCorporationSigned {
		return false
	}

	for _, a := range l.ReplacementAttorneys.Attorneys {
		if _, ok := replacementAttorneysSigned[a.ID]; !ok {
			return false
		}
	}

	if l.Attorneys.TrustCorporation.Name != "" && !trustCorporationSigned {
		return false
	}

	for _, a := range l.Attorneys.Attorneys {
		if _, ok := attorneysSigned[a.ID]; !ok {
			return false
		}
	}

	return true
}

type AddressDetail struct {
	Name    string
	Role    Type
	Address place.Address
	ID      string
}

func (l *DonorProvidedDetails) ActorAddresses() []place.Address {
	var addresses []place.Address

	if l.Donor.Address.String() != "" {
		addresses = append(addresses, l.Donor.Address)
	}

	if l.CertificateProvider.Address.String() != "" && !slices.Contains(addresses, l.CertificateProvider.Address) {
		addresses = append(addresses, l.CertificateProvider.Address)
	}

	for _, address := range l.Attorneys.Addresses() {
		if address.String() != "" && !slices.Contains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	for _, address := range l.ReplacementAttorneys.Addresses() {
		if address.String() != "" && !slices.Contains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	return addresses
}

func (l *DonorProvidedDetails) AllLayAttorneysFirstNames() []string {
	var names []string

	for _, a := range l.Attorneys.Attorneys {
		names = append(names, a.FirstNames)
	}

	for _, a := range l.ReplacementAttorneys.Attorneys {
		names = append(names, a.FirstNames)
	}

	return names
}

func (l *DonorProvidedDetails) AllLayAttorneysFullNames() []string {
	var names []string

	for _, a := range l.Attorneys.Attorneys {
		names = append(names, a.FullName())
	}

	for _, a := range l.ReplacementAttorneys.Attorneys {
		names = append(names, a.FullName())
	}

	return names
}

func (l *DonorProvidedDetails) TrustCorporationsNames() []string {
	var names []string

	if l.Attorneys.TrustCorporation.Name != "" {
		names = append(names, l.Attorneys.TrustCorporation.Name)
	}

	if l.ReplacementAttorneys.TrustCorporation.Name != "" {
		names = append(names, l.ReplacementAttorneys.TrustCorporation.Name)
	}

	return names
}

func (l *DonorProvidedDetails) Cost() int {
	if l.Tasks.PayForLpa.IsDenied() {
		return 8200
	}

	return pay.Cost(l.FeeType, l.PreviousFee)
}

func (l *DonorProvidedDetails) FeeAmount() int {
	paid := 0

	for _, payment := range l.PaymentDetails {
		paid += payment.Amount
	}

	return l.Cost() - paid
}

// CertificateProviderSharesDetails will return true if the last name or address
// of the certificate provider matches that of the donor or one of the
// attorneys. For a match of the last name we break on '-' to account for
// double-barrelled names.
func (l *DonorProvidedDetails) CertificateProviderSharesDetails() bool {
	certificateProviderParts := strings.Split(l.CertificateProvider.LastName, "-")

	donorParts := strings.Split(l.Donor.LastName, "-")
	for _, certificateProviderPart := range certificateProviderParts {
		if slices.Contains(donorParts, certificateProviderPart) {
			return true
		}

		if l.CertificateProvider.Address.Line1 == l.Donor.Address.Line1 &&
			l.CertificateProvider.Address.Postcode == l.Donor.Address.Postcode {
			return true
		}
	}

	for _, attorney := range append(l.Attorneys.Attorneys, l.ReplacementAttorneys.Attorneys...) {
		attorneyParts := strings.Split(attorney.LastName, "-")

		for _, certificateProviderPart := range certificateProviderParts {
			if slices.Contains(attorneyParts, certificateProviderPart) {
				return true
			}

			if l.CertificateProvider.Address.Line1 == attorney.Address.Line1 &&
				l.CertificateProvider.Address.Postcode == attorney.Address.Postcode {
				return true
			}
		}
	}

	return false
}
