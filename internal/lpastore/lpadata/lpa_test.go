package lpadata

import (
	"slices"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/stretchr/testify/assert"
)

func TestAllAttorneysSigned(t *testing.T) {
	attorneySigned := time.Now()

	testcases := map[string]struct {
		lpa      Lpa
		expected bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}, {}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{}, {SignedAt: attorneySigned}}},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}, {SignedAt: attorneySigned}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{SignedAt: attorneySigned}}},
			},
			expected: true,
		},
		"trust corporations not signed": {
			lpa: Lpa{
				Attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
			},
			expected: false,
		},
		"trust corporations signatory not signed": {
			lpa: Lpa{
				Attorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{}}}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}}}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: Lpa{
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}, {SignedAt: attorneySigned}}}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r", Signatories: []TrustCorporationSignatory{{SignedAt: attorneySigned}}}},
			},
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.AllAttorneysSigned())
		})
	}
}

func TestLpaCorrespondentEmail(t *testing.T) {
	lpa := &Lpa{
		Donor: Donor{Email: "donor"},
	}
	assert.Equal(t, "donor", lpa.CorrespondentEmail())
}

func TestLpaCorrespondentEmailWhenCorrespondentProvided(t *testing.T) {
	lpa := &Lpa{
		Donor:         Donor{Email: "donor"},
		Correspondent: Correspondent{Email: "correspondent"},
	}
	assert.Equal(t, "correspondent", lpa.CorrespondentEmail())
}

func TestLpaActors(t *testing.T) {
	lpa := &Lpa{
		Donor: Donor{
			UID:        actoruid.New(),
			FirstNames: "Sam",
			LastName:   "Smith",
		},
		CertificateProvider: CertificateProvider{
			UID:        actoruid.New(),
			FirstNames: "Charlie",
			LastName:   "Cooper",
		},
		Attorneys: Attorneys{
			Attorneys: []Attorney{{
				UID:        actoruid.New(),
				FirstNames: "Alan",
				LastName:   "Attorney",
			}, {
				UID:        actoruid.New(),
				FirstNames: "Angela",
				LastName:   "Attorney",
			}},
			TrustCorporation: TrustCorporation{Name: "Trusty"},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{{
				UID:        actoruid.New(),
				FirstNames: "Richard",
				LastName:   "Replacement",
			}, {
				UID:        actoruid.New(),
				FirstNames: "Rachel",
				LastName:   "Replacement",
			}},
			TrustCorporation: TrustCorporation{Name: "Untrusty"},
		},
		PeopleToNotify: []PersonToNotify{{
			UID:        actoruid.New(),
			FirstNames: "Peter",
			LastName:   "Person",
		}},
		AuthorisedSignatory: AuthorisedSignatory{
			UID:        actoruid.New(),
			FirstNames: "Aut",
			LastName:   "Sig",
		},
		IndependentWitness: IndependentWitness{
			UID:        actoruid.New(),
			FirstNames: "Ind",
			LastName:   "Wit",
		},
		Correspondent: Correspondent{FirstNames: "Nope"},
		Voucher:       Voucher{FirstNames: "Nada"},
	}

	actors := slices.Collect(lpa.Actors())

	assert.Equal(t, []actor.Actor{{
		Type:       actor.TypeDonor,
		UID:        lpa.Donor.UID,
		FirstNames: "Sam",
		LastName:   "Smith",
	}, {
		Type:       actor.TypeCertificateProvider,
		UID:        lpa.CertificateProvider.UID,
		FirstNames: "Charlie",
		LastName:   "Cooper",
	}, {
		Type:       actor.TypeAttorney,
		UID:        lpa.Attorneys.Attorneys[0].UID,
		FirstNames: "Alan",
		LastName:   "Attorney",
	}, {
		Type:       actor.TypeAttorney,
		UID:        lpa.Attorneys.Attorneys[1].UID,
		FirstNames: "Angela",
		LastName:   "Attorney",
	}, {
		Type:       actor.TypeReplacementAttorney,
		UID:        lpa.ReplacementAttorneys.Attorneys[0].UID,
		FirstNames: "Richard",
		LastName:   "Replacement",
	}, {
		Type:       actor.TypeReplacementAttorney,
		UID:        lpa.ReplacementAttorneys.Attorneys[1].UID,
		FirstNames: "Rachel",
		LastName:   "Replacement",
	}, {
		Type:       actor.TypePersonToNotify,
		UID:        lpa.PeopleToNotify[0].UID,
		FirstNames: "Peter",
		LastName:   "Person",
	}, {
		Type:       actor.TypeAuthorisedSignatory,
		UID:        lpa.AuthorisedSignatory.UID,
		FirstNames: "Aut",
		LastName:   "Sig",
	}, {
		Type:       actor.TypeIndependentWitness,
		UID:        lpa.IndependentWitness.UID,
		FirstNames: "Ind",
		LastName:   "Wit",
	}}, actors)
}

func TestAttorney(t *testing.T) {
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	lpa := &Lpa{
		Attorneys: Attorneys{
			Attorneys:        []Attorney{{UID: attorneyUID, FirstNames: "A", LastName: "B"}},
			TrustCorporation: TrustCorporation{UID: trustCorporationUID, Name: "C"},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys:        []Attorney{{UID: replacementAttorneyUID, FirstNames: "D", LastName: "E"}},
			TrustCorporation: TrustCorporation{UID: replacementTrustCorporationUID, Name: "F"},
		},
	}

	testcases := map[string]struct {
		uid       actoruid.UID
		name      string
		actorType actor.Type
	}{
		"attorney": {
			uid:       attorneyUID,
			name:      "A B",
			actorType: actor.TypeAttorney,
		},
		"replacement attorney": {
			uid:       replacementAttorneyUID,
			name:      "D E",
			actorType: actor.TypeReplacementAttorney,
		},
		"trust corporation": {
			uid:       trustCorporationUID,
			name:      "C",
			actorType: actor.TypeTrustCorporation,
		},
		"replacement trust corporation": {
			uid:       replacementTrustCorporationUID,
			name:      "F",
			actorType: actor.TypeReplacementTrustCorporation,
		},
		"missing": {
			uid:       actoruid.New(),
			actorType: actor.TypeNone,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			name, actorType := lpa.Attorney(tc.uid)

			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.actorType, actorType)
		})
	}
}
