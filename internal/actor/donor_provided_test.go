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

func TestCanBeUsedWhen(t *testing.T) {
	values := map[CanBeUsedWhen]string{CanBeUsedWhenCapacityLost: "when-capacity-lost", CanBeUsedWhenHasCapacity: "when-has-capacity"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseCanBeUsedWhen(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseCanBeUsedWhen("invalid")
		assert.NotNil(t, err)
	})
}

func TestLifeSustainingTreatment(t *testing.T) {
	values := map[LifeSustainingTreatment]string{LifeSustainingTreatmentOptionA: "option-a", LifeSustainingTreatmentOptionB: "option-b"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseLifeSustainingTreatment(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseLifeSustainingTreatment("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsOptionA", func(t *testing.T) {
		assert.True(t, LifeSustainingTreatmentOptionA.IsOptionA())
		assert.False(t, LifeSustainingTreatmentOptionB.IsOptionA())
	})

	t.Run("IsOptionB", func(t *testing.T) {
		assert.True(t, LifeSustainingTreatmentOptionB.IsOptionB())
		assert.False(t, LifeSustainingTreatmentOptionA.IsOptionB())
	})
}

func TestReplacementAttorneysStepIn(t *testing.T) {
	values := map[ReplacementAttorneysStepIn]string{
		ReplacementAttorneysStepInWhenAllCanNoLongerAct: "all",
		ReplacementAttorneysStepInWhenOneCanNoLongerAct: "one",
		ReplacementAttorneysStepInAnotherWay:            "other",
	}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseReplacementAttorneysStepIn(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseReplacementAttorneysStepIn("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsWhenAllCanNoLongerAct", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsWhenAllCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsWhenAllCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInAnotherWay.IsWhenAllCanNoLongerAct())
	})

	t.Run("IsWhenOneCanNoLongerAct", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsWhenOneCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsWhenOneCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInAnotherWay.IsWhenOneCanNoLongerAct())
	})

	t.Run("IsAnotherWay", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInAnotherWay.IsAnotherWay())
		assert.False(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsAnotherWay())
		assert.False(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsAnotherWay())
	})
}
