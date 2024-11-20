// Package donordata provides types that describe the data entered by a donor.
package donordata

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/mitchellh/hashstructure/v2"
)

const (
	currentHashVersion        uint8 = 0
	currentCheckedHashVersion uint8 = 0
)

type Tasks struct {
	YourDetails                task.State
	ChooseAttorneys            task.State
	ChooseReplacementAttorneys task.State
	WhenCanTheLpaBeUsed        task.State // property and affairs only
	LifeSustainingTreatment    task.State // personal welfare only
	Restrictions               task.State
	CertificateProvider        task.State
	PeopleToNotify             task.State
	AddCorrespondent           task.State
	ChooseYourSignatory        task.State // if .Donor.CanSign.IsNo only
	CheckYourLpa               task.State
	PayForLpa                  task.PaymentState
	ConfirmYourIdentity        task.IdentityState
	SignTheLpa                 task.State
}

// Provided contains all the data related to the LPA application
type Provided struct {
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
	Type lpadata.LpaType
	// Whether the applicant wants to add replacement attorneys
	WantReplacementAttorneys form.YesNo
	// When the LPA can be used
	WhenCanTheLpaBeUsed lpadata.CanBeUsedWhen
	// Preferences on life sustaining treatment (applicable to personal welfare LPAs only)
	LifeSustainingTreatmentOption lpadata.LifeSustainingTreatment
	// Restrictions on attorneys actions
	Restrictions string
	// Used to show the task list
	Tasks Tasks
	// PaymentDetails are records of payments made for the LPA via GOV.UK Pay
	PaymentDetails []Payment
	// Information returned by the identity service related to the Donor or Voucher
	IdentityUserData identity.UserData
	// Replacement attorneys named in the LPA
	ReplacementAttorneys Attorneys
	// Information on how the applicant wishes their replacement attorneys to act
	ReplacementAttorneyDecisions AttorneyDecisions
	// How to bring in replacement attorneys, if set
	HowShouldReplacementAttorneysStepIn lpadata.ReplacementAttorneysStepIn
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
	// StatutoryWaitingPeriodAt is when the Lpa transitioned to the STATUTORY_WAITING_PERIOD
	// status in the lpa-store
	StatutoryWaitingPeriodAt time.Time
	// RegisteringWithCourtOfProtection is set when the donor wishes to take the
	// Lpa to the Court of Protection for registration.
	RegisteringWithCourtOfProtection bool
	// Version is the number of times the LPA has been updated (auto-incremented
	// on PUT)
	Version int `hash:"-"`

	// WantVoucher indicates if the donor knows someone who can vouch for them and wants
	// then to do so
	WantVoucher form.YesNo
	// Voucher is a person the donor has nominated to vouch for their identity
	Voucher Voucher
	// FailedVouchAttempts are the number of unsuccessful attempts a voucher has made to confirm the Donors ID
	FailedVouchAttempts int

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

// CanChange returns true if the donor can make changes to their LPA.
func (p *Provided) CanChange() bool {
	return p.SignedAt.IsZero()
}

func (p *Provided) HashInclude(field string, _ any) (bool, error) {
	if p.HashVersion > currentHashVersion {
		return false, errors.New("HashVersion too high")
	}

	return true, nil
}

// toCheck filters the fields used for hashing further, for the use of
// determining whether the LPA data has changed since it was checked by the
// donor.
type toCheck Provided

func (c toCheck) HashInclude(field string, _ any) (bool, error) {
	if c.CheckedHashVersion > currentCheckedHashVersion {
		return false, errors.New("CheckedHashVersion too high")
	}

	// The following fields don't contain LPA data, so aren't part of what gets
	// checked.
	switch field {
	case "CheckedAt",
		"CreatedAt",
		"Tasks",
		"PaymentDetails",
		"IdentityUserData",
		"WantToApplyForLpa",
		"WantToSignLpa",
		"SignedAt",
		"SubmittedAt",
		"WithdrawnAt",
		"StatutoryWaitingPeriodAt",
		"CertificateProviderCodes",
		"WitnessedByCertificateProviderAt",
		"IndependentWitnessCodes",
		"WitnessedByIndependentWitnessAt",
		"WitnessCodeLimiter",
		"FeeType",
		"EvidenceDelivery",
		"PreviousApplicationNumber",
		"PreviousFee",
		"RegisteringWithCourtOfProtection",
		"WantVoucher",
		"Voucher",
		"FailedVouchAttempts":
		return false, nil
	}

	return true, nil
}

func (p *Provided) NamesChanged(firstNames, lastName, otherNames string) bool {
	return p.Donor.FirstNames != firstNames || p.Donor.LastName != lastName || p.Donor.OtherNames != otherNames
}

func (p *Provided) HashChanged() bool {
	hash, _ := p.generateHash()

	return hash != p.Hash
}

func (p *Provided) UpdateHash() (err error) {
	p.HashVersion = currentHashVersion
	p.Hash, err = p.generateHash()
	return err
}

func (p *Provided) generateHash() (uint64, error) {
	return hashstructure.Hash(p, hashstructure.FormatV2, nil)
}

func (p *Provided) CheckedHashChanged() bool {
	hash, _ := p.generateCheckedHash()

	return hash != p.CheckedHash
}

func (p *Provided) UpdateCheckedHash() (err error) {
	p.CheckedHashVersion = currentCheckedHashVersion
	p.CheckedHash, err = p.generateCheckedHash()
	return err
}

func (p *Provided) generateCheckedHash() (uint64, error) {
	return hashstructure.Hash(toCheck(*p), hashstructure.FormatV2, nil)
}

func (p *Provided) DonorIdentityConfirmed() bool {
	return p.IdentityUserData.Status.IsConfirmed() &&
		p.IdentityUserData.MatchName(p.Donor.FirstNames, p.Donor.LastName) &&
		p.IdentityUserData.DateOfBirth.Equals(p.Donor.DateOfBirth)
}

// SigningDeadline gives the date at which the LPA should be signed by the
// certificate provider and attorneys.
func (p *Provided) SigningDeadline() time.Time {
	if p.RegisteringWithCourtOfProtection {
		return p.SignedAt.AddDate(0, 4, 14)
	}

	return p.SignedAt.AddDate(0, 0, 28)
}

// IdentityDeadline gives the date which the donor must complete their identity
// confirmation, otherwise the signature will expire.
func (p *Provided) IdentityDeadline() time.Time {
	if p.WitnessedByCertificateProviderAt.IsZero() {
		return time.Time{}
	}

	return p.WitnessedByCertificateProviderAt.AddDate(0, 6, 0)
}

// CourtOfProtectionSubmissionDeadline gives the date at which the signed LPA
// must be submitted to the Court of Protection, if registering through this
// route.
func (p *Provided) CourtOfProtectionSubmissionDeadline() time.Time {
	return p.SignedAt.AddDate(0, 6, 0)
}

type Under18ActorDetails struct {
	FullName    string
	DateOfBirth date.Date
	UID         actoruid.UID
	Type        actor.Type
}

func (p *Provided) Under18ActorDetails() []Under18ActorDetails {
	var data []Under18ActorDetails
	eighteenYearsAgo := date.Today().AddDate(-18, 0, 0)

	for _, a := range p.Attorneys.Attorneys {
		if a.DateOfBirth.After(eighteenYearsAgo) {
			data = append(data, Under18ActorDetails{
				FullName:    a.FullName(),
				DateOfBirth: a.DateOfBirth,
				UID:         a.UID,
				Type:        actor.TypeAttorney,
			})
		}
	}

	for _, ra := range p.ReplacementAttorneys.Attorneys {
		if ra.DateOfBirth.After(eighteenYearsAgo) {
			data = append(data, Under18ActorDetails{
				FullName:    ra.FullName(),
				DateOfBirth: ra.DateOfBirth,
				UID:         ra.UID,
				Type:        actor.TypeReplacementAttorney,
			})
		}
	}

	return data
}

func (p *Provided) CorrespondentEmail() string {
	if p.Correspondent.Email == "" {
		return p.Donor.Email
	}

	return p.Correspondent.Email
}

func (p *Provided) ActorAddresses() []place.Address {
	var addresses []place.Address

	if p.Donor.Address.String() != "" {
		addresses = append(addresses, p.Donor.Address)
	}

	if p.CertificateProvider.Address.String() != "" && !slices.Contains(addresses, p.CertificateProvider.Address) {
		addresses = append(addresses, p.CertificateProvider.Address)
	}

	for _, address := range p.Attorneys.Addresses() {
		if address.String() != "" && !slices.Contains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	for _, address := range p.ReplacementAttorneys.Addresses() {
		if address.String() != "" && !slices.Contains(addresses, address) {
			addresses = append(addresses, address)
		}
	}

	return addresses
}

func (p *Provided) AllLayAttorneysFirstNames() []string {
	var names []string

	for _, a := range p.Attorneys.Attorneys {
		names = append(names, a.FirstNames)
	}

	for _, a := range p.ReplacementAttorneys.Attorneys {
		names = append(names, a.FirstNames)
	}

	return names
}

func (p *Provided) AllLayAttorneysFullNames() []string {
	var names []string

	for _, a := range p.Attorneys.Attorneys {
		names = append(names, a.FullName())
	}

	for _, a := range p.ReplacementAttorneys.Attorneys {
		names = append(names, a.FullName())
	}

	return names
}

func (p *Provided) TrustCorporationsNames() []string {
	var names []string

	if p.Attorneys.TrustCorporation.Name != "" {
		names = append(names, p.Attorneys.TrustCorporation.Name)
	}

	if p.ReplacementAttorneys.TrustCorporation.Name != "" {
		names = append(names, p.ReplacementAttorneys.TrustCorporation.Name)
	}

	return names
}

func (p *Provided) Cost() int {
	if p.Tasks.PayForLpa.IsDenied() {
		return 8200
	}

	return pay.Cost(p.FeeType, p.PreviousFee)
}

func (p *Provided) FeeAmount() pay.AmountPence {
	paid := 0

	for _, payment := range p.PaymentDetails {
		paid += payment.Amount
	}

	return pay.AmountPence(p.Cost() - paid)
}

// CertificateProviderSharesDetails will return true if the last name or address
// of the certificate provider matches that of the donor or one of the
// attorneys. For a match of the last name we break on '-' to account for
// double-barrelled names.
func (p *Provided) CertificateProviderSharesDetails() bool {
	certificateProviderParts := strings.Split(p.CertificateProvider.LastName, "-")

	donorParts := strings.Split(p.Donor.LastName, "-")
	for _, certificateProviderPart := range certificateProviderParts {
		if slices.Contains(donorParts, certificateProviderPart) {
			return true
		}

		if p.CertificateProvider.Address.Line1 == p.Donor.Address.Line1 &&
			p.CertificateProvider.Address.Postcode == p.Donor.Address.Postcode {
			return true
		}
	}

	for _, attorney := range append(p.Attorneys.Attorneys, p.ReplacementAttorneys.Attorneys...) {
		attorneyParts := strings.Split(attorney.LastName, "-")

		for _, certificateProviderPart := range certificateProviderParts {
			if slices.Contains(attorneyParts, certificateProviderPart) {
				return true
			}

			if p.CertificateProvider.Address.Line1 == attorney.Address.Line1 &&
				p.CertificateProvider.Address.Postcode == attorney.Address.Postcode {
				return true
			}
		}
	}

	return false
}

// Actors returns an iterator over all human actors named on the LPA (i.e. this
// excludes trust corporations, the correspondent, and the voucher).
func (p *Provided) Actors() iter.Seq[actor.Actor] {
	return func(yield func(actor.Actor) bool) {
		if !yield(actor.Actor{
			Type:       actor.TypeDonor,
			UID:        p.Donor.UID,
			FirstNames: p.Donor.FirstNames,
			LastName:   p.Donor.LastName,
		}) {
			return
		}

		if !yield(actor.Actor{
			Type:       actor.TypeCertificateProvider,
			UID:        p.CertificateProvider.UID,
			FirstNames: p.CertificateProvider.FirstNames,
			LastName:   p.CertificateProvider.LastName,
		}) {
			return
		}

		for _, attorney := range p.Attorneys.Attorneys {
			if !yield(actor.Actor{
				Type:       actor.TypeAttorney,
				UID:        attorney.UID,
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
			}) {
				return
			}
		}

		for _, attorney := range p.ReplacementAttorneys.Attorneys {
			if !yield(actor.Actor{
				Type:       actor.TypeReplacementAttorney,
				UID:        attorney.UID,
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
			}) {
				return
			}
		}

		for _, person := range p.PeopleToNotify {
			if !yield(actor.Actor{
				Type:       actor.TypePersonToNotify,
				UID:        person.UID,
				FirstNames: person.FirstNames,
				LastName:   person.LastName,
			}) {
				return
			}
		}

		if p.AuthorisedSignatory.FirstNames != "" {
			if !yield(actor.Actor{
				Type:       actor.TypeAuthorisedSignatory,
				FirstNames: p.AuthorisedSignatory.FirstNames,
				LastName:   p.AuthorisedSignatory.LastName,
			}) {
				return
			}
		}

		if p.IndependentWitness.FirstNames != "" {
			if !yield(actor.Actor{
				Type:       actor.TypeIndependentWitness,
				FirstNames: p.IndependentWitness.FirstNames,
				LastName:   p.IndependentWitness.LastName,
			}) {
				return
			}
		}
	}
}

func (p *Provided) CanHaveVoucher() bool {
	return p.FailedVouchAttempts < 2
}

func (p *Provided) UpdateDecisions() {
	if p.Attorneys.Len() <= 1 {
		p.AttorneyDecisions = AttorneyDecisions{}
	} else {
		if !p.AttorneyDecisions.How.IsJointlyForSomeSeverallyForOthers() {
			p.AttorneyDecisions.Details = ""
		}
	}

	if p.ReplacementAttorneys.Len() <= 1 {
		p.ReplacementAttorneyDecisions = AttorneyDecisions{}
	} else {
		if !p.ReplacementAttorneyDecisions.How.IsJointlyForSomeSeverallyForOthers() {
			p.ReplacementAttorneyDecisions.Details = ""
		}

		if p.Attorneys.Len() == 1 || p.AttorneyDecisions.How.IsJointly() {
			p.HowShouldReplacementAttorneysStepIn = lpadata.ReplacementAttorneysStepIn(0)
		} else if p.AttorneyDecisions.How.IsJointlyAndSeverally() {
			if p.ReplacementAttorneys.Len() <= 1 || !p.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() {
				p.ReplacementAttorneyDecisions = AttorneyDecisions{}
			}
		} else {
			p.ReplacementAttorneyDecisions = AttorneyDecisions{}
			p.HowShouldReplacementAttorneysStepIn = lpadata.ReplacementAttorneysStepIn(0)
		}
	}
}
