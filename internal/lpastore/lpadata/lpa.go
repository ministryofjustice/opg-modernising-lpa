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
	Submitted                                bool
	Paid                                     bool
	IsOrganisationDonor                      bool
	Drafted                                  bool
	CannotRegister                           bool
	Correspondent                            Correspondent
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
