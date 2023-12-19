package actor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaType(t *testing.T) {
	values := map[LpaType]string{LpaTypePersonalWelfare: "personal-welfare", LpaTypePropertyAndAffairs: "property-and-affairs"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse %s", s), func(t *testing.T) {
			parsed, err := ParseLpaType(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string %s", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseLpaType("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsHealthWelfare", func(t *testing.T) {
		assert.True(t, LpaTypePersonalWelfare.IsHealthWelfare())
		assert.False(t, LpaTypePropertyAndAffairs.IsHealthWelfare())
	})

	t.Run("IsPropertyFinance", func(t *testing.T) {
		assert.True(t, LpaTypePropertyAndAffairs.IsPropertyFinance())
		assert.False(t, LpaTypePersonalWelfare.IsPropertyFinance())
	})
}

func TestTypeLegalTermTransKey(t *testing.T) {
	testCases := map[LpaType]string{
		LpaTypePropertyAndAffairs: "pfaLegalTerm",
		LpaTypePersonalWelfare:    "hwLegalTerm",
		LpaType(99):               "",
		LpaType(0):                "",
	}

	for lpaType, translationKey := range testCases {
		t.Run(lpaType.String(), func(t *testing.T) {
			assert.Equal(t, translationKey, lpaType.LegalTermTransKey())
		})
	}
}

func TestTypeWhatLPACoversTransKey(t *testing.T) {
	testCases := map[LpaType]string{
		LpaTypePropertyAndAffairs: "whatPropertyAndAffairsCovers",
		LpaTypePersonalWelfare:    "whatPersonalWelfareCovers",
		LpaType(99):               "",
		LpaType(0):                "",
	}

	for lpaType, translationKey := range testCases {
		t.Run(lpaType.String(), func(t *testing.T) {
			assert.Equal(t, translationKey, lpaType.WhatLPACoversTransKey())
		})
	}
}
