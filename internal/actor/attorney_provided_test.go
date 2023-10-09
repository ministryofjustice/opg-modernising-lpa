package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyProvidedDetailsSigned(t *testing.T) {
	certificateProviderSignedAt := time.Now()
	signedAfter := certificateProviderSignedAt.Add(time.Second)

	testcases := map[string]struct {
		details AttorneyProvidedDetails
		signed  bool
	}{
		"unsigned": {},
		"signed": {
			details: AttorneyProvidedDetails{Confirmed: signedAfter},
			signed:  true,
		},
		"signed before": {
			details: AttorneyProvidedDetails{Confirmed: certificateProviderSignedAt},
		},
		"trust corporation unsigned": {
			details: AttorneyProvidedDetails{Confirmed: signedAfter, IsTrustCorporation: true},
		},
		"trust corporation single signatory": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.No,
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{Confirmed: signedAfter}, {}},
			},
			signed: true,
		},
		"trust corporation signle signatory unsigned": {
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
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{Confirmed: signedAfter}, {Confirmed: signedAfter}},
			},
			signed: true,
		},
		"trust corporation double signatory unsigned": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{Confirmed: signedAfter}, {}},
			},
		},
		"trust corporation double signatory signed before": {
			details: AttorneyProvidedDetails{
				IsTrustCorporation:       true,
				WouldLikeSecondSignatory: form.Yes,
				AuthorisedSignatories:    [2]TrustCorporationSignatory{{Confirmed: certificateProviderSignedAt}, {Confirmed: signedAfter}},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.signed, tc.details.Signed(certificateProviderSignedAt))
		})
	}
}

func TestAttorneyProvidedDetailsSignedWhenPreviousActorNotSigned(t *testing.T) {
	assert.False(t, AttorneyProvidedDetails{Confirmed: time.Now()}.Signed(time.Time{}))
}
