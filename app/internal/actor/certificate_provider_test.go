package actor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"

	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderFullName(t *testing.T) {
	p := CertificateProvider{FirstNames: "Bob Alan George", LastName: "Smith Jones-Doe"}

	assert.Equal(t, "Bob Alan George Smith Jones-Doe", p.FullName())
}

func TestCertificateProviderIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		cp       *CertificateProvider
		expected bool
	}{
		"set": {
			cp: &CertificateProvider{
				FirstNames:       "a",
				LastName:         "b",
				IdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"missing provider": {
			cp: &CertificateProvider{
				IdentityUserData: identity.UserData{OK: true},
			},
			expected: false,
		},
		"not ok": {
			cp: &CertificateProvider{
				IdentityUserData: identity.UserData{Provider: identity.OneLogin},
			},
			expected: false,
		},
		"no match": {
			cp: &CertificateProvider{
				FirstNames:       "a",
				LastName:         "b",
				IdentityUserData: identity.UserData{Provider: identity.OneLogin},
			},
			expected: false,
		},
		"none": {
			cp:       &CertificateProvider{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cp.CertificateProviderIdentityConfirmed())
		})
	}
}
