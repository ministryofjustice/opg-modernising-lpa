package random

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomAlphaNumeric(t *testing.T) {
	for _, length := range []int{1, 10, 100, 999} {
		got := AlphaNumeric(length)

		assert.Len(t, got, length)
	}
}

func TestRandomFriendly(t *testing.T) {
	for _, length := range []int{1, 10, 100, 999} {
		got := Friendly(length)

		assert.Regexp(t, fmt.Sprintf("^[346789BCDFGHJKMPQRTVWXY]{%d}$", length), got)
	}
}

func TestRandomNumeric(t *testing.T) {
	for _, length := range []int{1, 10, 100, 999} {
		got := Numeric(length)

		assert.Regexp(t, fmt.Sprintf("^[0-9]{%d}$", length), got)
	}
}
