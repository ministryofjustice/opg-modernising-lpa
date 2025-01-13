package lpadata

import (
	"iter"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type Lpa struct {
	LpaKey                                     dynamo.LpaKeyType
	LpaOwnerKey                                dynamo.LpaOwnerKeyType
	LpaID                                      string
	LpaUID                                     string
	Status                                     Status
	RegisteredAt                               time.Time
	WithdrawnAt                                time.Time
	StatutoryWaitingPeriodAt                   time.Time
	UpdatedAt                                  time.Time
	Type                                       LpaType
	Donor                                      Donor
	Attorneys                                  Attorneys
	ReplacementAttorneys                       Attorneys
	CertificateProvider                        CertificateProvider
	PeopleToNotify                             []PersonToNotify
	AttorneyDecisions                          AttorneyDecisions
	ReplacementAttorneyDecisions               AttorneyDecisions
	HowShouldReplacementAttorneysStepIn        ReplacementAttorneysStepIn
	HowShouldReplacementAttorneysStepInDetails string
	Restrictions                               string
	WhenCanTheLpaBeUsed                        CanBeUsedWhen
	LifeSustainingTreatmentOption              LifeSustainingTreatment
	AuthorisedSignatory                        AuthorisedSignatory
	IndependentWitness                         IndependentWitness

	// SignedAt is when the donor signed their LPA.
	SignedAt time.Time

	// WitnessedByCertificateProviderAt is when the certificate provider signed to
	// say they witnessed the donor signing the LPA.
	WitnessedByCertificateProviderAt time.Time

	// WitnessedByIndependentWitnessAt is when the independent witness signed to
	// say they witnessed the LPA being signed on the donor's behalf, if the donor
	// said they were unable to sign themselves.
	WitnessedByIndependentWitnessAt time.Time

	// CertificateProviderNotRelatedConfirmedAt is when the donor confirmed that
	// the certificate provider is not related to them, if they have similar
	// details.
	CertificateProviderNotRelatedConfirmedAt time.Time

	// Submitted is set if SubmittedAt is non-zero for online applications, or set
	// to true for paper applications.
	Submitted bool

	// Paid is set if the PayForLpa task has been completed for online
	// applications, or set to true for paper applications as to be in the
	// lpa-store the application payment must be complete.
	Paid bool

	// IsOrganisationDonor is set to true when the Lpa is being made by a
	// supporter working for an organisation.
	IsOrganisationDonor bool

	// Drafted is set if the CheckYourLpa task has been completed for online
	// applications, or set to true for paper applications.
	Drafted bool

	// Correspondent is set using the data provided by the donor for online
	// applications, but is not set for paper applications.
	Correspondent Correspondent

	// Voucher is set using the data provided by the donor for online
	// applications, but is not set for paper applications.
	Voucher Voucher

	// CertificateProviderInvitedAt is when the certificate provider's share
	// code is first sent, it is only set with the resolving service.
	CertificateProviderInvitedAt time.Time

	// AttorneysInvitedAt records when the share codes are sent to the attorneys,
	// it is only set with the resolving service.
	AttorneysInvitedAt time.Time

	// InStore is true if the Lpa is in the lpa-store.
	InStore bool
}

// SignedForDonor returns true if the Lpa has been signed and witnessed for the donor.
func (l *Lpa) SignedForDonor() bool {
	return !l.SignedAt.IsZero() && !l.WitnessedByCertificateProviderAt.IsZero() &&
		(l.IndependentWitness.FirstNames == "" || !l.WitnessedByIndependentWitnessAt.IsZero())
}

func (l *Lpa) CorrespondentEmail() string {
	if l.Correspondent.Email == "" {
		return l.Donor.Email
	}

	return l.Correspondent.Email
}

func (l Lpa) AllAttorneysSigned() bool {
	if l.Attorneys.Len() == 0 {
		return false
	}

	for _, attorneys := range []Attorneys{l.Attorneys, l.ReplacementAttorneys} {
		for _, a := range attorneys.Attorneys {
			if a.SignedAt == nil || a.SignedAt.IsZero() {
				return false
			}
		}

		if t := attorneys.TrustCorporation; t.Name != "" {
			if len(t.Signatories) == 0 {
				return false
			}

			for _, s := range t.Signatories {
				if s.SignedAt.IsZero() {
					return false
				}
			}
		}
	}

	return true
}

// Actors returns an iterator over all human actors named on the LPA (i.e. this
// excludes trust corporations, the correspondent, and the voucher).
func (l Lpa) Actors() iter.Seq[actor.Actor] {
	return func(yield func(actor.Actor) bool) {
		if !yield(actor.Actor{
			Type:       actor.TypeDonor,
			UID:        l.Donor.UID,
			FirstNames: l.Donor.FirstNames,
			LastName:   l.Donor.LastName,
		}) {
			return
		}

		if !yield(actor.Actor{
			Type:       actor.TypeCertificateProvider,
			UID:        l.CertificateProvider.UID,
			FirstNames: l.CertificateProvider.FirstNames,
			LastName:   l.CertificateProvider.LastName,
		}) {
			return
		}

		for _, attorney := range l.Attorneys.Attorneys {
			if !yield(actor.Actor{
				Type:       actor.TypeAttorney,
				UID:        attorney.UID,
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
			}) {
				return
			}
		}

		for _, attorney := range l.ReplacementAttorneys.Attorneys {
			if !yield(actor.Actor{
				Type:       actor.TypeReplacementAttorney,
				UID:        attorney.UID,
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
			}) {
				return
			}
		}

		for _, person := range l.PeopleToNotify {
			if !yield(actor.Actor{
				Type:       actor.TypePersonToNotify,
				UID:        person.UID,
				FirstNames: person.FirstNames,
				LastName:   person.LastName,
			}) {
				return
			}
		}

		if !l.AuthorisedSignatory.UID.IsZero() {
			if !yield(actor.Actor{
				Type:       actor.TypeAuthorisedSignatory,
				UID:        l.AuthorisedSignatory.UID,
				FirstNames: l.AuthorisedSignatory.FirstNames,
				LastName:   l.AuthorisedSignatory.LastName,
			}) {
				return
			}
		}

		if !l.IndependentWitness.UID.IsZero() {
			if !yield(actor.Actor{
				Type:       actor.TypeIndependentWitness,
				UID:        l.IndependentWitness.UID,
				FirstNames: l.IndependentWitness.FirstNames,
				LastName:   l.IndependentWitness.LastName,
			}) {
				return
			}
		}
	}
}

func (l *Lpa) Attorney(uid actoruid.UID) (string, string, actor.Type) {
	if t := l.ReplacementAttorneys.TrustCorporation; t.UID == uid {
		return t.Name, t.Mobile, actor.TypeReplacementTrustCorporation
	}

	if t := l.Attorneys.TrustCorporation; t.UID == uid {
		return t.Name, t.Mobile, actor.TypeTrustCorporation
	}

	if a, ok := l.ReplacementAttorneys.Get(uid); ok {
		return a.FullName(), a.Mobile, actor.TypeReplacementAttorney
	}

	if a, ok := l.Attorneys.Get(uid); ok {
		return a.FullName(), a.Mobile, actor.TypeAttorney
	}

	return "", "", actor.TypeNone
}

// ExpiresAt gives the date the LPA expires.
func (l *Lpa) ExpiresAt() time.Time {
	if l.Submitted && l.Donor.IdentityCheck != nil && !l.Donor.IdentityCheck.CheckedAt.IsZero() {
		return l.SignedAt.AddDate(2, 0, 0)
	}

	if !l.SignedAt.IsZero() {
		return l.SignedAt.AddDate(0, 6, 0)
	}

	return time.Time{}
}
