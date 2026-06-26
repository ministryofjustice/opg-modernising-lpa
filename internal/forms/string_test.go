package forms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString_SetInput(t *testing.T) {
	s := NewString("a", "A")
	s.SetInput("  ok  ")

	assert.Equal(t, "ok", s.Input)
}
