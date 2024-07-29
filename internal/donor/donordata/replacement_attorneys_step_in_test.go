package donordata

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplacementAttorneysStepIn(t *testing.T) {
	values := map[ReplacementAttorneysStepIn]string{
		ReplacementAttorneysStepInWhenAllCanNoLongerAct: "all-can-no-longer-act",
		ReplacementAttorneysStepInWhenOneCanNoLongerAct: "one-can-no-longer-act",
		ReplacementAttorneysStepInAnotherWay:            "another-way",
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
