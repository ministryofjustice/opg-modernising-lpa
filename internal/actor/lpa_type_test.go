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
	testCases := map[string]struct {
		LpaType           LpaType
		ExpectedLegalTerm string
	}{
		"PFA": {
			LpaType:           LpaTypePropertyFinance,
			ExpectedLegalTerm: "pfaLegalTerm",
		},
		"HW": {
			LpaType:           LpaTypeHealthWelfare,
			ExpectedLegalTerm: "hwLegalTerm",
		},
		"unexpected": {
			LpaType:           LpaType(5),
			ExpectedLegalTerm: "",
		},
		"empty": {
			ExpectedLegalTerm: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedLegalTerm, tc.LpaType.LegalTermTransKey())
		})
	}
}

func TestTypeWhatLPACoversTransKey(t *testing.T) {
	testCases := map[string]struct {
		LpaType            LpaType
		ExpectedWhatCovers string
	}{
		"PFA": {
			LpaType:            LpaTypePropertyFinance,
			ExpectedWhatCovers: "whatPFACovers",
		},
		"HW": {
			LpaType:            LpaTypeHealthWelfare,
			ExpectedWhatCovers: "whatHWCovers",
		},
		"unexpected": {
			LpaType:            LpaType(5),
			ExpectedWhatCovers: "",
		},
		"empty": {
			ExpectedWhatCovers: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedWhatCovers, tc.LpaType.WhatLPACoversTransKey())
		})
	}
}
