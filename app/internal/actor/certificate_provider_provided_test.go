package actor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderProvidedFullName(t *testing.T) {
	p := CertificateProviderProvidedDetails{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", p.FullName())
}

func TestCertificateProviderProvidedIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		cp       *CertificateProviderProvidedDetails
		expected bool
	}{
		"set": {
			cp: &CertificateProviderProvidedDetails{
				FirstNames:       "a",
				LastName:         "b",
				IdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "a", LastName: "b"},
			},
			expected: true,
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
				FirstNames:       "a",
				LastName:         "b",
				IdentityUserData: identity.UserData{Provider: identity.OneLogin},
			},
			expected: false,
		},
		"none": {
			cp:       &CertificateProviderProvidedDetails{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cp.CertificateProviderIdentityConfirmed())
		})
	}
}
