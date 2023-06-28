package actor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYesNo(t *testing.T) {
	values := map[YesNo]string{Yes: "yes", No: "no"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseYesNo(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseYesNo("invalid")
		assert.NotNil(t, err)
	})
}
