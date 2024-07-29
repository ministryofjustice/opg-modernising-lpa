package certificateproviderdata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderProvidedIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		cp          *Provided
		firstNames  string
		lastName    string
		dateOfBirth date.Date
		expected    bool
	}{
		"confirmed": {
			cp: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusConfirmed, DateOfBirth: date.New("2000", "1", "1")},
				DateOfBirth:      date.New("2000", "1", "1"),
			},
			firstNames:  "a",
			lastName:    "b",
			dateOfBirth: date.New("2000", "1", "1"),
			expected:    true,
		},
		"failed": {
			cp: &Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusFailed},
			},
			expected: false,
		},
		"name does not match": {
			cp: &Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			firstNames: "a",
			lastName:   "c",
			expected:   false,
		},
		"dob does not match": {
			cp: &Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
			firstNames:  "a",
			lastName:    "b",
			dateOfBirth: date.New("2000", "1", "1"),
			expected:    false,
		},
		"insufficient evidence": {
			cp: &Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			},
			firstNames: "a",
			lastName:   "b",
			expected:   false,
		},
		"none": {
			cp:       &Provided{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cp.CertificateProviderIdentityConfirmed(tc.firstNames, tc.lastName))
		})
	}
}
