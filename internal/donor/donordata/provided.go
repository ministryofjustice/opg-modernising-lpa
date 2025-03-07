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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/mitchellh/hashstructure/v2"
)

type Localizer interface {
	localize.Localizer
}

const (
	currentHashVersion                                       uint8 = 0
	currentCheckedHashVersion                                uint8 = 0
	currentCertificateProviderNotRelatedConfirmedHashVersion uint8 = 0
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
	PK dynamo.LpaKeyType      `hash:"-" checkhash:"-"`
	SK dynamo.LpaOwnerKeyType `hash:"-" checkhash:"-"`
	// Hash is used to determine whether the Lpa has been changed since last read
	Hash uint64 `hash:"-" checkhash:"-"`
	// HashVersion is used to determine the fields used to calculate Hash
	HashVersion uint8 `hash:"-" checkhash:"-"`
	// LpaID identifies the LPA being drafted
	LpaID string
	// LpaUID is a unique identifier created after sending basic LPA details to the UID service
	LpaUID string `dynamodbav:",omitempty"`
	// CreatedAt is when the LPA was created
	CreatedAt time.Time `checkhash:"-"`
	// UpdatedAt is when the LPA was last updated
	UpdatedAt time.Time `hash:"-" checkhash:"-"`
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
	Tasks Tasks `checkhash:"-"`
	// PaymentDetails are records of payments made for the LPA via GOV.UK Pay
	PaymentDetails []Payment `checkhash:"-"`
	// Information returned by the identity service related to the Donor or Voucher
	IdentityUserData identity.UserData `checkhash:"-"`
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
	WantToApplyForLpa bool `checkhash:"-"`
	// Confirmation that the applicant wants to sign the LPA
	WantToSignLpa bool `checkhash:"-"`
	// CertificateProviderNotRelatedConfirmedAt is when the donor confirmed the
	// certificate provider is not related to another similar actor
	CertificateProviderNotRelatedConfirmedAt time.Time
	// CertificateProviderNotRelatedConfirmedHash is the hash of data that was confirmed by
	// CertificateProviderNotRelatedConfirmedAt
	CertificateProviderNotRelatedConfirmedHash uint64
	// CertificateProviderNotRelatedConfirmedHashVersion is used to determine the
	// fields used to calculate CertificateProviderNotRelatedConfirmedHash
	CertificateProviderNotRelatedConfirmedHashVersion uint8
	// CheckedAt is when the donor checked their LPA
	CheckedAt time.Time `checkhash:"-"`
	// CheckedHash is the Hash value of the LPA when last checked
	CheckedHash uint64 `hash:"-" checkhash:"-"`
	// CheckedHashVersion is used to determine the fields used to calculate CheckedHash
	CheckedHashVersion uint8 `hash:"-" checkhash:"-"`
	// SignedAt is when the donor submitted their signature
	SignedAt time.Time `checkhash:"-"`
	// WithdrawnAt is when the Lpa was withdrawn by the donor
	WithdrawnAt time.Time `checkhash:"-"`
	// StatutoryWaitingPeriodAt is when the Lpa transitioned to the
	// statutory-waiting-period status in the lpa-store
	StatutoryWaitingPeriodAt time.Time `checkhash:"-"`
	// DoNotRegisterAt is when the Lpa transitioned to do-not-register status in
	// the lpa-store
	DoNotRegisterAt time.Time `checkhash:"-"`
	// RegisteringWithCourtOfProtection is set when the donor wishes to take the
	// Lpa to the Court of Protection for registration.
	RegisteringWithCourtOfProtection bool `checkhash:"-"`
	// ContinueWithMismatchedIdentity is set when the donor wishes to continue
	// their application with mismatched identity details
	ContinueWithMismatchedIdentity bool `checkhash:"-"`
	// Version is the number of times the LPA has been updated (auto-incremented
	// on PUT)
	Version int `hash:"-" checkhash:"-"`

	// WantVoucher indicates if the donor knows someone who can vouch for them and wants
	// then to do so
	WantVoucher form.YesNo `checkhash:"-"`
	// Voucher is a person the donor has nominated to vouch for their identity
	Voucher Voucher `checkhash:"-"`
	// VouchAttempts are the number of attempts a voucher has made to confirm the Donors identity
	VouchAttempts int `checkhash:"-"`
	// FailedVoucher is the last voucher that was unable to vouch for the donor
	FailedVoucher Voucher `checkhash:"-"`

	// Codes used for the certificate provider to witness signing
	CertificateProviderCodes WitnessCodes `checkhash:"-"`
	// When the signing was witnessed by the certificate provider
	WitnessedByCertificateProviderAt time.Time `checkhash:"-"`
	// Codes used for the independent witness to witness signing
	IndependentWitnessCodes WitnessCodes `checkhash:"-"`
	// When the signing was witnessed by the independent witness
	WitnessedByIndependentWitnessAt time.Time `checkhash:"-"`
	// Used to rate limit witness code attempts
	WitnessCodeLimiter *Limiter `checkhash:"-"`

	// FeeType is the type of fee the user is applying for
	FeeType pay.FeeType `checkhash:"-"`
	// EvidenceDelivery is the method by which the user wants to send evidence
	EvidenceDelivery pay.EvidenceDelivery `checkhash:"-"`
	// PreviousApplicationNumber if the application is related to an existing application
	PreviousApplicationNumber string `checkhash:"-"`
	// PreviousFee is the fee previously paid for an LPA, if applying for a repeat
	// of an LPA with reference prefixed 7 or have selected HalfFee for
	// CostOfRepeatApplication.
	PreviousFee pay.PreviousFee `checkhash:"-"`
	// CostOfRepeatApplication is the fee the donor believes they are eligible
	// for, if applying for a repeat of an LPA with reference prefixed M.
	CostOfRepeatApplication pay.CostOfRepeatApplication `checkhash:"-"`

	// CertificateProviderInvitedAt records when the invite is sent to the
	// certificate provider to act.
	CertificateProviderInvitedAt time.Time `checkhash:"-"`

	// AttorneysInvitedAt records when the invites are sent to the attorneys.
	AttorneysInvitedAt time.Time `checkhash:"-"`

	// VoucherInvitedAt records when the invite is sent to the voucher to vouch.
	VoucherInvitedAt time.Time `checkhash:"-"`

	// DetailsVerifiedByVoucher records that a voucher has verified details supplied by
	// the donor match their identity.
	DetailsVerifiedByVoucher bool `checkhash:"-"`

	// MoreEvidenceRequiredAt records when a request for further information on an
	// exemption/remission was received.
	MoreEvidenceRequiredAt time.Time `checkhash:"-"`

	// PriorityCorrespondenceSentAt records when a caseworker sent a letter to the
	// donor informing them of a problem.
	PriorityCorrespondenceSentAt time.Time `checkhash:"-"`

	// MaterialChangeConfirmedAt records when a material change to LPA data was
	// confirmed by a caseworker
	MaterialChangeConfirmedAt time.Time `checkhash:"-"`

	// ImmaterialChangeConfirmedAt records when an immaterial change to LPA data was
	// confirmed by a caseworker
	ImmaterialChangeConfirmedAt time.Time `checkhash:"-"`

	// HasSeenSuccessfulVouchBanner records if the donor has seen the progress
	// tracker successful vouch banner
	HasSeenSuccessfulVouchBanner bool `checkhash:"-"`

	// HasSeenReducedFeeApprovalNotification records if the donor has seen the
	// progress tracker exemption/remission fee approved banner
	HasSeenReducedFeeApprovalNotification bool `checkhash:"-"`

	// HasSeenIdentityMismatchResolvedNotification records if the donor has seen
	// the progress tracker identity confirmed banner
	HasSeenIdentityMismatchResolvedNotification bool `checkhash:"-"`

	// HasSeenCertificateProviderIdentityMismatchResolvedNotification records if
	// the donor has seen the progress tracker certificate provider identity
	// confirmed banner
	HasSeenCertificateProviderIdentityMismatchResolvedNotification bool `checkhash:"-"`

	// ReducedFeeApprovedAt records when an exemption/remission was approved.
	ReducedFeeApprovedAt time.Time `checkhash:"-"`

	// IdentityDetailsCausedCheck is set when details are updated to match
	// confirmed identity, and check and send hasn't been done with those new
	// details
	IdentityDetailsCausedCheck bool `checkhash:"-"`

	HasSentApplicationUpdatedEvent bool `hash:"-" checkhash:"-"`
}

func (p *Provided) CompletedAllTasks() bool {
	return p.Tasks.YourDetails.IsCompleted() &&
		p.Tasks.ChooseAttorneys.IsCompleted() &&
		p.Tasks.ChooseReplacementAttorneys.IsCompleted() &&
		(p.Type.IsPropertyAndAffairs() && p.Tasks.WhenCanTheLpaBeUsed.IsCompleted() ||
			p.Type.IsPersonalWelfare() && p.Tasks.LifeSustainingTreatment.IsCompleted()) &&
		p.Tasks.Restrictions.IsCompleted() &&
		p.Tasks.CertificateProvider.IsCompleted() &&
		p.Tasks.PeopleToNotify.IsCompleted() &&
		p.Tasks.AddCorrespondent.IsCompleted() &&
		(p.Donor.CanSign.IsYes() || p.Tasks.ChooseYourSignatory.IsCompleted()) &&
		p.Tasks.CheckYourLpa.IsCompleted() &&
		p.Tasks.PayForLpa.IsCompleted() &&
		p.Tasks.ConfirmYourIdentity.IsCompleted() &&
		p.Tasks.SignTheLpa.IsCompleted()
}

// CanChange returns true if the donor can make changes to their LPA.
func (p *Provided) CanChange() bool {
	return p.SignedAt.IsZero()
}

// CanChangePersonalDetails returns true if the donor can make changes to their FirstNames, LastName or DateOfBirth.
func (p *Provided) CanChangePersonalDetails() bool {
	return !p.IdentityUserData.Status.IsConfirmed() &&
		p.SignedAt.IsZero() &&
		!p.DetailsVerifiedByVoucher
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

	return true, nil
}

// toConfirmCertificateProviderNotRelated filters the fields used for hashing to
// only those used in CertificateProviderSharesDetails
type toConfirmCertificateProviderNotRelated Provided

func (c toConfirmCertificateProviderNotRelated) HashInclude(field string, _ any) (bool, error) {
	if c.CertificateProviderNotRelatedConfirmedHashVersion > currentCertificateProviderNotRelatedConfirmedHashVersion {
		return false, errors.New("CertificateProviderNotRelatedConfirmedHashVersion too high")
	}

	return field == "CertificateProvider" || field == "Donor" || field == "Attorneys" || field == "ReplacementAttorneys", nil
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

// UpdateCheckedHash will generate a value that can be compared to check if any
// fields containing LPA data have changed. Fields that do not contain LPA data,
// so should be ignored for this calculation, are tagged with `checkhash:"-"`.
func (p *Provided) UpdateCheckedHash() (err error) {
	p.CheckedHashVersion = currentCheckedHashVersion
	p.CheckedHash, err = p.generateCheckedHash()
	return err
}

func (p *Provided) generateCheckedHash() (uint64, error) {
	return hashstructure.Hash(toCheck(*p), hashstructure.FormatV2, &hashstructure.HashOptions{TagName: "checkhash"})
}

func (p *Provided) CertificateProviderNotRelatedConfirmedHashChanged() bool {
	hash, _ := p.generateCertificateProviderNotRelatedConfirmedHash()

	return hash != p.CertificateProviderNotRelatedConfirmedHash
}

// UpdateCertificateProviderNotRelatedConfirmedHash will generate a value that
// can be compared to check if any fields containing LPA data have
// changed. Fields that do not contain LPA data, so should be ignored for this
// calculation, are tagged with `relatedhash:"-"`.
func (p *Provided) UpdateCertificateProviderNotRelatedConfirmedHash() (err error) {
	p.CertificateProviderNotRelatedConfirmedHashVersion = currentCertificateProviderNotRelatedConfirmedHashVersion
	p.CertificateProviderNotRelatedConfirmedHash, err = p.generateCertificateProviderNotRelatedConfirmedHash()
	return err
}

func (p *Provided) generateCertificateProviderNotRelatedConfirmedHash() (uint64, error) {
	return hashstructure.Hash(toConfirmCertificateProviderNotRelated(*p), hashstructure.FormatV2, &hashstructure.HashOptions{TagName: "relatedhash"})
}

func (p *Provided) DonorIdentityConfirmed() bool {
	return p.IdentityUserData.Status.IsConfirmed() &&
		(p.IdentityUserData.MatchName(p.Donor.FirstNames, p.Donor.LastName) &&
			p.IdentityUserData.DateOfBirth.Equals(p.Donor.DateOfBirth) ||
			p.ContinueWithMismatchedIdentity && !p.ImmaterialChangeConfirmedAt.IsZero())
}

// SignatoriesNames returns the full names of the non-donor actors expected to
// sign the LPA.
func (p *Provided) SignatoriesNames(localizer Localizer) []string {
	return append([]string{p.CertificateProvider.FullName()}, p.AttorneysNames(localizer)...)
}

// AttorneysNames returns the full names of the attorneys and trust corporation.
func (p *Provided) AttorneysNames(localizer Localizer) []string {
	var names []string

	if p.HasTrustCorporation() {
		names = append(names, localizer.Format("aSignatoryFromTrustCorporation", map[string]any{
			"TrustCorporationName": p.TrustCorporation().Name,
		}))
	}

	return append(names, p.AllLayAttorneysFullNames()...)
}

// SigningDeadline gives the date at which the LPA should be signed by the
// certificate provider and attorneys.
func (p *Provided) SigningDeadline() time.Time {
	if p.RegisteringWithCourtOfProtection {
		return p.SignedAt.AddDate(0, 4, 14)
	}

	return p.SignedAt.AddDate(2, 0, 0)
}

// DonorSigningDeadline gives the date at which the LPA should be signed by the
// donor once identity is confirmed.
func (p *Provided) DonorSigningDeadline() time.Time {
	if !p.IdentityUserData.Status.IsConfirmed() {
		return time.Time{}
	}

	return p.IdentityUserData.CheckedAt.AddDate(0, 6, 0)
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

// CertificateProviderDeadline gives the date at which the certificate provider
// should act.
func (p *Provided) CertificateProviderDeadline() time.Time {
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

func (p *Provided) HasTrustCorporation() bool {
	return p.Attorneys.TrustCorporation.Name != "" || p.ReplacementAttorneys.TrustCorporation.Name != ""
}

func (p *Provided) TrustCorporation() TrustCorporation {
	if p.Attorneys.TrustCorporation.Name != "" {
		return p.Attorneys.TrustCorporation
	}

	return p.ReplacementAttorneys.TrustCorporation
}

func (p *Provided) Cost() int {
	if p.Tasks.PayForLpa.IsDenied() {
		return 8200
	}

	return pay.Cost(p.FeeType, p.PreviousFee, p.CostOfRepeatApplication)
}

func (p *Provided) Paid() pay.AmountPence {
	var paid pay.AmountPence
	for _, payment := range p.PaymentDetails {
		paid += pay.AmountPence(payment.Amount)
	}

	return paid
}

func (p *Provided) FeeAmount() pay.AmountPence {
	return pay.AmountPence(p.Cost()) - p.Paid()
}

// PaidAt returns the latest date a payment was made.
func (p *Provided) PaidAt() time.Time {
	var at time.Time
	for _, payment := range p.PaymentDetails {
		if payment.CreatedAt.After(at) {
			at = payment.CreatedAt
		}
	}

	return at
}

// CertificateProviderSharesDetails will return true if the last name or address
// of the certificate provider matches that of the donor or one of the
// attorneys. For a match of the last name we break on '-' to account for
// double-barrelled names.
func (p *Provided) CertificateProviderSharesDetails() bool {
	if !p.CertificateProviderNotRelatedConfirmedAt.IsZero() && !p.CertificateProviderNotRelatedConfirmedHashChanged() {
		return false
	}

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
	return p.VouchAttempts < 2
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
