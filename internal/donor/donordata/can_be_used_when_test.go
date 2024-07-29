package donordata

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
