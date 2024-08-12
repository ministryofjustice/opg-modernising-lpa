package lpadata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type Lpa struct {
	LpaKey                                     dynamo.LpaKeyType
	LpaOwnerKey                                dynamo.LpaOwnerKeyType
	LpaID                                      string
	LpaUID                                     string
	RegisteredAt                               time.Time
	WithdrawnAt                                time.Time
	PerfectAt                                  time.Time
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

	// SignedAt is the date the Donor signed their LPA (and signifies it has been
	// witnessed by their CertificateProvider)
	SignedAt                                 time.Time
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

	// CannotRegister is set to true if the status in the lpa-store is
	// cannot-register.
	CannotRegister bool

	// Correspondent is set using the data set by the donor for online
	// applications, but is not set for paper applications.
	Correspondent Correspondent
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
			if a.SignedAt.IsZero() {
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
