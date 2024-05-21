package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyProvidedDetailsSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		details AttorneyProvidedDetails
		signed  bool
	}{
		"unsigned": {},
		"signed": {
			details: AttorneyProvidedDetails{Confirmed: attorneySignedAt},
			signed:  true,
		},
		"trust corporation unsigned": {
			details: AttorneyProvidedDetails{Confirmed: attorneySignedAt, IsTrustCorporation: true},
		},
		"trust corporation single signatory": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{Confirmed: attorneySignedAt},
					{},
				},
			},
			signed: true,
		},
		"trust corporation single signatory unsigned": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{}, {}},
			},
		},
		"trust corporation double signatory": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{Confirmed: attorneySignedAt},
					{Confirmed: attorneySignedAt},
				},
			},
			signed: true,
		},
		"trust corporation double signatory unsigned": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{Confirmed: attorneySignedAt},
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
