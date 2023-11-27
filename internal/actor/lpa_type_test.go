package actor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLpaType(t *testing.T) {
	values := map[LpaType]string{LpaTypeHealthWelfare: "hw", LpaTypePropertyFinance: "pfa"}

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
		assert.True(t, LpaTypeHealthWelfare.IsHealthWelfare())
		assert.False(t, LpaTypePropertyFinance.IsHealthWelfare())
	})

	t.Run("IsPropertyFinance", func(t *testing.T) {
		assert.True(t, LpaTypePropertyFinance.IsPropertyFinance())
		assert.False(t, LpaTypeHealthWelfare.IsPropertyFinance())
	})
}

func TestTypeLegalTermTransKey(t *testing.T) {
	testCases := map[LpaType]string{
		LpaTypePropertyFinance: "pfaLegalTerm",
		LpaTypeHealthWelfare:   "hwLegalTerm",
		LpaType(99):            "",
		LpaType(0):             "",
	}

	for lpaType, translationKey := range testCases {
		t.Run(lpaType.String(), func(t *testing.T) {
			assert.Equal(t, translationKey, lpaType.LegalTermTransKey())
		})
	}
}

func TestTypeWhatLPACoversTransKey(t *testing.T) {
	testCases := map[LpaType]string{
		LpaTypePropertyFinance: "whatPersonalAffairsCovers",
		LpaTypeHealthWelfare:   "whatPersonalWelfareCovers",
		LpaType(99):            "",
		LpaType(0):             "",
	}

	for lpaType, translationKey := range testCases {
		t.Run(lpaType.String(), func(t *testing.T) {
			assert.Equal(t, translationKey, lpaType.WhatLPACoversTransKey())
		})
	}
}
