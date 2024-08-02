package lpadata

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

	t.Run("IsPersonalWelfare", func(t *testing.T) {
		assert.True(t, LpaTypePersonalWelfare.IsPersonalWelfare())
		assert.False(t, LpaTypePropertyAndAffairs.IsPersonalWelfare())
	})

	t.Run("IsPropertyAndAffairs", func(t *testing.T) {
		assert.True(t, LpaTypePropertyAndAffairs.IsPropertyAndAffairs())
		assert.False(t, LpaTypePersonalWelfare.IsPropertyAndAffairs())
	})
}

func TestLpaTypeWhatLPACoversTransKey(t *testing.T) {
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
