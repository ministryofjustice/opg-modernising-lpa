package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmountPencePound(t *testing.T) {
	assert.Equal(t, "£15", AmountPence(1500).String())
	assert.Equal(t, "£103.27", AmountPence(10327).String())
	assert.Equal(t, "£945,678.99", AmountPence(94567899).String())
}

func TestAmountPenceInt(t *testing.T) {
	assert.Equal(t, 123, AmountPence(123).Pence())
}
