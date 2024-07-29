package donordata

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
