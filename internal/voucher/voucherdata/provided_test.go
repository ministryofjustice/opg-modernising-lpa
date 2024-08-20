package voucherdata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
)

func TestProvidedFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Provided{FirstNames: "John", LastName: "Smith"}.FullName())
}

func TestProvidedIdentityConfirmed(t *testing.T) {
	assert.True(t, Provided{
		FirstNames: "X",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusConfirmed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}.IdentityConfirmed())
}

func TestProvidedIdentityConfirmedWhenNameNotMatch(t *testing.T) {
	assert.False(t, Provided{
		FirstNames: "A",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusConfirmed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}.IdentityConfirmed())
}

func TestProvidedIdentityConfirmedWhenNotConfirmed(t *testing.T) {
	assert.False(t, Provided{
		FirstNames: "X",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusFailed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}.IdentityConfirmed())
}
