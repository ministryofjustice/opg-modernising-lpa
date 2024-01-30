package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	for _, length := range []int{1, 10, 100, 999} {
		got := String(length)

		assert.Len(t, got, length)
	}
}

func TestRandomCode(t *testing.T) {
	for _, length := range []int{1, 10, 100, 999} {
		got := Code(length)

		assert.Len(t, got, length)
	}
}
