package actor

import (
	"testing"

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
				IdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b"},
			},
			firstNames: "a",
			lastName:   "b",
			expected:   true,
		},
		"not ok": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{},
			},
			expected: false,
		},
		"no match": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{},
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
