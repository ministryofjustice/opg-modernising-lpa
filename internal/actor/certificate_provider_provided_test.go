package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderProvidedIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		cp         *CertificateProviderProvidedDetails
		firstNames string
		lastName   string
		expected   bool
	}{
		"set": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "a", LastName: "b"},
			},
			firstNames: "a",
			lastName:   "b",
			expected:   true,
		},
		"missing provider": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{OK: true},
			},
			expected: false,
		},
		"not ok": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Provider: identity.OneLogin},
			},
			expected: false,
		},
		"no match": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Provider: identity.OneLogin},
			},
			firstNames: "a",
			lastName:   "b",
			expected:   false,
		},
		"none": {
			cp:       &CertificateProviderProvidedDetails{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cp.CertificateProviderIdentityConfirmed(tc.firstNames, tc.lastName))
		})
	}
}

func TestCertificateProviderProvidedSigned(t *testing.T) {
	now := time.Now()
	assert := assert.New(t)

	assert.False(CertificateProviderProvidedDetails{}.Signed(now))
	assert.False(CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: now}}.Signed(now))
	assert.True(CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: now.Add(time.Second)}}.Signed(now))
}
