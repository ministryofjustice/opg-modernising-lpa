package lpadata

import (
	"testing"
	"time"

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