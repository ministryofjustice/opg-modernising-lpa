package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderFullName(t *testing.T) {
	assert.Equal(t, "First Last", CertificateProvider{FirstNames: "First", LastName: "Last"}.FullName())
}
