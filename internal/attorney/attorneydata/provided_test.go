package attorneydata

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestProvidedSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		details Provided
		signed  bool
	}{
		"unsigned": {},
		"signed": {
			details: Provided{SignedAt: attorneySignedAt},
			signed:  true,
		},
		"trust corporation unsigned": {
			details: Provided{SignedAt: attorneySignedAt, IsTrustCorporation: true},
		},
		"trust corporation single signatory": {
			details: Provided{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{SignedAt: attorneySignedAt},
					{},
				},
			},
			signed: true,
		},
		"trust corporation single signatory unsigned": {
			details: Provided{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{}, {}},
			},
		},
		"trust corporation double signatory": {
			details: Provided{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{SignedAt: attorneySignedAt},
					{SignedAt: attorneySignedAt},
				},
			},
			signed: true,
		},
		"trust corporation double signatory unsigned": {
			details: Provided{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{SignedAt: attorneySignedAt},
					{},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.signed, tc.details.Signed())
		})
	}
}
