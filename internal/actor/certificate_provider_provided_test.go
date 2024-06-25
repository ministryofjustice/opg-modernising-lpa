package actor

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderProvidedIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		cp          *CertificateProviderProvidedDetails
		firstNames  string
		lastName    string
		dateOfBirth date.Date
		expected    bool
	}{
		"confirmed": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.IdentityStatusConfirmed, DateOfBirth: date.New("2000", "1", "1")},
				DateOfBirth:      date.New("2000", "1", "1"),
			},
			firstNames:  "a",
			lastName:    "b",
			dateOfBirth: date.New("2000", "1", "1"),
			expected:    true,
		},
		"failed": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Status: identity.IdentityStatusFailed},
			},
			expected: false,
		},
		"name does not match": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Status: identity.IdentityStatusConfirmed},
			},
			firstNames: "a",
			lastName:   "c",
			expected:   false,
		},
		"dob does not match": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Status: identity.IdentityStatusConfirmed},
			},
			firstNames:  "a",
			lastName:    "b",
			dateOfBirth: date.New("2000", "1", "1"),
			expected:    false,
		},
		"insufficient evidence": {
			cp: &CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{Status: identity.IdentityStatusInsufficientEvidence},
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
