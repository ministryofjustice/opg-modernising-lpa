package voucherdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvidedFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Provided{FirstNames: "John", LastName: "Smith"}.FullName())
}
