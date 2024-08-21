package voucherdata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestProvidedFullName(t *testing.T) {
	provided := &Provided{FirstNames: "John", LastName: "Smith"}

	assert.Equal(t, "John Smith", provided.FullName())
}

func TestProvidedIdentityConfirmed(t *testing.T) {
	provided := &Provided{
		FirstNames: "X",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusConfirmed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}

	assert.True(t, provided.IdentityConfirmed())
}

func TestProvidedIdentityConfirmedWhenNameNotMatch(t *testing.T) {
	provided := &Provided{
		FirstNames: "A",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusConfirmed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}

	assert.False(t, provided.IdentityConfirmed())
}

func TestProvidedIdentityConfirmedWhenNotConfirmed(t *testing.T) {
	provided := &Provided{
		FirstNames: "X",
		LastName:   "Y",
		IdentityUserData: identity.UserData{
			Status:     identity.StatusFailed,
			FirstNames: "X",
			LastName:   "Y",
		},
	}

	assert.False(t, provided.IdentityConfirmed())
}

func TestProvidedNameMatches(t *testing.T) {
	provided := &Provided{FirstNames: "A", LastName: "B"}
	lpa := &lpadata.Lpa{Voucher: lpadata.Voucher{FirstNames: "A", LastName: "B"}}

	assert.Equal(t, actor.TypeNone, provided.NameMatches(lpa))
}

func TestProvidedNameMatchesWhenUnset(t *testing.T) {
	provided := &Provided{}
	lpa := &lpadata.Lpa{}

	assert.Equal(t, actor.TypeNone, provided.NameMatches(lpa))
}

func TestProvidedNameMatchesWhenMatch(t *testing.T) {
	provided := &Provided{FirstNames: "A", LastName: "B"}
	lpa := &lpadata.Lpa{Donor: lpadata.Donor{FirstNames: "A", LastName: "B"}}

	assert.Equal(t, actor.TypeDonor, provided.NameMatches(lpa))
}
