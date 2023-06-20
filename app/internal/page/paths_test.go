package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathString(t *testing.T) {
	assert.Equal(t, "/anything", Path("/anything").String())
}

func TestPathFormat(t *testing.T) {
	assert.Equal(t, "/anything", Path("/anything").Format())
}

func TestLpaPathString(t *testing.T) {
	assert.Equal(t, "/anything", LpaPath("/anything").String())
}

func TestLpaPathFormat(t *testing.T) {
	assert.Equal(t, "/lpa/abc/anything", LpaPath("/anything").Format("abc"))
}

func TestAttorneyPathString(t *testing.T) {
	assert.Equal(t, "/anything", AttorneyPath("/anything").String())
}

func TestAttorneyPathFormat(t *testing.T) {
	assert.Equal(t, "/attorney/abc/anything", AttorneyPath("/anything").Format("abc"))
}

func TestCertificateProviderPathString(t *testing.T) {
	assert.Equal(t, "/anything", CertificateProviderPath("/anything").String())
}

func TestCertificateProviderPathFormat(t *testing.T) {
	assert.Equal(t, "/certificate-provider/abc/anything", CertificateProviderPath("/anything").Format("abc"))
}
