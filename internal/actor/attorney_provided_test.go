package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyProvidedDetailsSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		details AttorneyProvidedDetails
		signed  bool
	}{
		"unsigned": {},
		"signed": {
			details: AttorneyProvidedDetails{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
			signed:  true,
		},
		"signed for different iteration": {
			details: AttorneyProvidedDetails{LpaSignedAt: otherLpaSignedAt, Confirmed: lpaSignedAt},
		},
		"trust corporation unsigned": {
			details: AttorneyProvidedDetails{Confirmed: attorneySignedAt, IsTrustCorporation: true},
		},
		"trust corporation single signatory": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
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
					{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
					{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
				},
			},
			signed: true,
		},
		"trust corporation double signatory unsigned": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
					{},
				},
			},
		},
		"trust corporation double signatory signed before": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories: [2]TrustCorporationSignatory{
					{LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySignedAt},
					{LpaSignedAt: lpaSignedAt, Confirmed: attorneySignedAt},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.signed, tc.details.Signed(lpaSignedAt))
		})
	}
}

func TestAttorneyProvidedDetailsSignedWhenLpaNotSigned(t *testing.T) {
	assert.False(t, AttorneyProvidedDetails{Confirmed: time.Now()}.Signed(time.Time{}))
}
