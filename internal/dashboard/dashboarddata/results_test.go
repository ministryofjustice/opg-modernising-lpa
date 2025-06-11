package dashboarddata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestResultsEmpty(t *testing.T) {
	results := Results{}
	assert.True(t, results.Empty())

	results.Donor = append(results.Donor, Actor{})
	assert.False(t, results.Empty())
}

func TestResultsByActorType(t *testing.T) {
	donor := []Actor{{Lpa: &lpadata.Lpa{LpaID: "donor"}}}
	attorney := []Actor{{Lpa: &lpadata.Lpa{LpaID: "attorney"}}}
	certificateProvider := []Actor{{Lpa: &lpadata.Lpa{LpaID: "cp"}}}
	voucher := []Actor{{Lpa: &lpadata.Lpa{LpaID: "voucher"}}}

	results := Results{Donor: donor, Attorney: attorney, CertificateProvider: certificateProvider, Voucher: voucher}

	assert.Equal(t, donor, results.ByActorType(actor.TypeDonor))

	assert.Equal(t, certificateProvider, results.ByActorType(actor.TypeCertificateProvider))

	assert.Equal(t, attorney, results.ByActorType(actor.TypeAttorney))
	assert.Equal(t, attorney, results.ByActorType(actor.TypeReplacementAttorney))
	assert.Equal(t, attorney, results.ByActorType(actor.TypeTrustCorporation))
	assert.Equal(t, attorney, results.ByActorType(actor.TypeReplacementTrustCorporation))

	assert.Equal(t, voucher, results.ByActorType(actor.TypeVoucher))

	assert.Equal(t, []Actor{}, results.ByActorType(actor.TypeCorrespondent))
}
