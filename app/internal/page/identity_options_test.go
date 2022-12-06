package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadIdentityOption(t *testing.T) {
	assert.Equal(t, Passport, readIdentityOption("passport"))
	assert.Equal(t, IdentityOptionUnknown, readIdentityOption("what"))
}

func TestIdentityOptionArticleLabel(t *testing.T) {
	assert.Equal(t, "theYoti", Yoti.ArticleLabel())
	assert.Equal(t, "", IdentityOptionUnknown.ArticleLabel())
}

func TestIdentityOptionLabel(t *testing.T) {
	assert.Equal(t, "yoti", Yoti.Label())
	assert.Equal(t, "", IdentityOptionUnknown.Label())
}

func TestIdentityOptionNextPath(t *testing.T) {
	options := IdentityOptions{First: Yoti, Second: GovernmentGatewayAccount}

	assert.Equal(t, appData.Paths.IdentityWithYoti, options.NextPath(IdentityOptionUnknown, appData.Paths))
	assert.Equal(t, appData.Paths.IdentityWithGovernmentGatewayAccount, options.NextPath(Yoti, appData.Paths))
	assert.Equal(t, appData.Paths.WhatHappensWhenSigning, options.NextPath(GovernmentGatewayAccount, appData.Paths))
}

func TestIdentityOptionsRanked(t *testing.T) {
	first, second := identityOptionsRanked([]IdentityOption{Passport, DrivingLicence, DwpAccount})
	assert.Equal(t, first, Passport)
	assert.Equal(t, second, DrivingLicence)

	first, second = identityOptionsRanked([]IdentityOption{GovernmentGatewayAccount, CouncilTaxBill, UtilityBill})
	assert.Equal(t, first, GovernmentGatewayAccount)
	assert.Equal(t, second, UtilityBill)

	first, second = identityOptionsRanked([]IdentityOption{DrivingLicence, OnlineBankAccount, GovernmentGatewayAccount})
	assert.Equal(t, first, DrivingLicence)
	assert.Equal(t, second, GovernmentGatewayAccount)

	first, second = identityOptionsRanked([]IdentityOption{})
	assert.Equal(t, first, IdentityOptionUnknown)
	assert.Equal(t, second, IdentityOptionUnknown)

	first, second = identityOptionsRanked([]IdentityOption{Yoti, Passport, DrivingLicence})
	assert.Equal(t, first, Yoti)
	assert.Equal(t, second, IdentityOptionUnknown)
}
