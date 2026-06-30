package forms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_SetInput(t *testing.T) {
	s := NewString("a", "A")
	s.Set("  ok  ")

	assert.Equal(t, "ok", s.Input)
}
