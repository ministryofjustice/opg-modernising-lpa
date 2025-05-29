package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderFullName(t *testing.T) {
	assert.Equal(t, "First Last", CertificateProvider{FirstNames: "First", LastName: "Last"}.FullName())
}

func TestCertificateProviderNameHasChanged(t *testing.T) {
	assert.False(t, CertificateProvider{FirstNames: "a", LastName: "b"}.NameHasChanged("a", "b"))
	assert.True(t, CertificateProvider{FirstNames: "a", LastName: "b"}.NameHasChanged("a", ""))
}

func TestCertificateProviderAddressHasChanged(t *testing.T) {
	assert.False(t, CertificateProvider{Address: testAddress}.AddressHasChanged(testAddress))
	assert.True(t, CertificateProvider{Address: testAddress}.AddressHasChanged(place.Address{Line1: "a"}))
}
