package actor

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/mitchellh/hashstructure/v2"
)

const (
	currentHashVersion        uint8 = 0
	currentCheckedHashVersion uint8 = 0
)

type DonorTasks struct {
	YourDetails                TaskState
	ChooseAttorneys            TaskState
	ChooseReplacementAttorneys TaskState
	WhenCanTheLpaBeUsed        TaskState // property and affairs only
	LifeSustainingTreatment    TaskState // personal welfare only
	Restrictions               TaskState
	CertificateProvider        TaskState
	PeopleToNotify             TaskState
	AddCorrespondent           TaskState
	ChooseYourSignatory        TaskState // if .Donor.CanSign.IsNo only
	CheckYourLpa               TaskState
	PayForLpa                  PaymentTask
	ConfirmYourIdentityAndSign TaskState
}

// DonorProvidedDetails contains all the data related to the LPA application
type DonorProvidedDetails struct {
	PK dynamo.LpaKeyType      `hash:"-"`
	SK dynamo.LpaOwnerKeyType `hash:"-"`
	// Hash is used to determine whether the Lpa has been changed since last read
	Hash uint64 `hash:"-"`
	// HashVersion is used to determine the fields used to calculate Hash
	HashVersion uint8 `hash:"-"`
	// LpaID identifies the LPA being drafted
	LpaID string
	// LpaUID is a unique identifier created after sending basic LPA details to the UID service
	LpaUID string `dynamodbav:",omitempty"`
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
	// Whether the applicant wants to add a correspondent
	AddCorrespondent form.YesNo
	// Correspondent is sent updates on an application in place of a (supporter) donor
	Correspondent Correspondent
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
	// CertificateProviderNotRelatedConfirmedAt is when the donor confirmed the
	// certificate provider is not related to another similar actor
	CertificateProviderNotRelatedConfirmedAt time.Time
	// CheckedAt is when the donor checked their LPA
	CheckedAt time.Time
	// CheckedHash is the Hash value of the LPA when last checked
	CheckedHash uint64 `hash:"-"`
	// CheckedHashVersion is used to determine the fields used to calculate CheckedHash
	CheckedHashVersion uint8 `hash:"-"`
	// SignedAt is when the donor submitted their signature
	SignedAt time.Time
	// SubmittedAt is when the Lpa was sent to the OPG
	SubmittedAt time.Time
	// WithdrawnAt is when the Lpa was withdrawn by the donor
	WithdrawnAt time.Time
	// PerfectAt is when the Lpa transitioned to the PERFECT status in the lpa-store
	PerfectAt time.Time
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

	HasSentApplicationUpdatedEvent bool `hash:"-"`
}

func (d *DonorProvidedDetails) HashInclude(field string, _ any) (bool, error) {
	if d.HashVersion > currentHashVersion {
		return false, errors.New("HashVersion too high")
	}

	return true, nil
}

// toCheck filters the fields used for hashing further, for the use of
// determining whether the LPA data has changed since it was checked by the
// donor.
type toCheck DonorProvidedDetails

func (c toCheck) HashInclude(field string, _ any) (bool, error) {
	if c.CheckedHashVersion > currentCheckedHashVersion {
		return false, errors.New("CheckedHashVersion too high")
	}

	// The following fields don't contain LPA data, so aren't part of what gets
	// checked.
	switch field {
	case "CheckedAt",
		"Tasks",
		"PaymentDetails",
		"DonorIdentityUserData",
		"WantToApplyForLpa",
		"WantToSignLpa",
		"SignedAt",
		"SubmittedAt",
		"WithdrawnAt",
		"PerfectAt",
		"CertificateProviderCodes",
		"WitnessedByCertificateProviderAt",
		"IndependentWitnessCodes",
		"WitnessedByIndependentWitnessAt",
		"WitnessCodeLimiter",
		"FeeType",
		"EvidenceDelivery",
		"PreviousApplicationNumber",
		"PreviousFee":
		return false, nil
	}

	return true, nil
}

func (l *DonorProvidedDetails) NamesChanged(firstNames, lastName, otherNames string) bool {
	return l.Donor.FirstNames != firstNames || l.Donor.LastName != lastName || l.Donor.OtherNames != otherNames
}

func (l *DonorProvidedDetails) HashChanged() bool {
	hash, _ := l.generateHash()

	return hash != l.Hash
}

func (l *DonorProvidedDetails) UpdateHash() (err error) {
	l.HashVersion = currentHashVersion
	l.Hash, err = l.generateHash()
	return err
}

func (l *DonorProvidedDetails) generateHash() (uint64, error) {
	return hashstructure.Hash(l, hashstructure.FormatV2, nil)
}

func (l *DonorProvidedDetails) CheckedHashChanged() bool {
	hash, _ := l.generateCheckedHash()

	return hash != l.CheckedHash
}

func (l *DonorProvidedDetails) UpdateCheckedHash() (err error) {
	l.CheckedHashVersion = currentCheckedHashVersion
	l.CheckedHash, err = l.generateCheckedHash()
	return err
}

func (l *DonorProvidedDetails) generateCheckedHash() (uint64, error) {
	return hashstructure.Hash(toCheck(*l), hashstructure.FormatV2, nil)
}

func (l *DonorProvidedDetails) DonorIdentityConfirmed() bool {
	return l.DonorIdentityUserData.OK &&
		l.DonorIdentityUserData.MatchName(l.Donor.FirstNames, l.Donor.LastName) &&
		l.DonorIdentityUserData.DateOfBirth.Equals(l.Donor.DateOfBirth)
}

func (l *DonorProvidedDetails) AttorneysAndCpSigningDeadline() time.Time {
	return l.SignedAt.Add((24 * time.Hour) * 28)
}

type Under18ActorDetails struct {
	FullName    string
	DateOfBirth date.Date
	UID         actoruid.UID
	Type        Type
}

func (l *DonorProvidedDetails) Under18ActorDetails() []Under18ActorDetails {
	var data []Under18ActorDetails
	eighteenYearsAgo := date.Today().AddDate(-18, 0, 0)

	for _, a := range l.Attorneys.Attorneys {
		if a.DateOfBirth.After(eighteenYearsAgo) {
			data = append(data, Under18ActorDetails{
				FullName:    a.FullName(),
				DateOfBirth: a.DateOfBirth,
				UID:         a.UID,
				Type:        TypeAttorney,
			})
		}
	}

	for _, ra := range l.ReplacementAttorneys.Attorneys {
		if ra.DateOfBirth.After(eighteenYearsAgo) {
			data = append(data, Under18ActorDetails{
				FullName:    ra.FullName(),
				DateOfBirth: ra.DateOfBirth,
				UID:         ra.UID,
				Type:        TypeReplacementAttorney,
			})
		}
	}

	return data
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
