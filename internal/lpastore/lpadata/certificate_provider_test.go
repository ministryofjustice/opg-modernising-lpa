package lpadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderFullName(t *testing.T) {
	assert.Equal(t, "John Smith", CertificateProvider{FirstNames: "John", LastName: "Smith"}.FullName())
}
