package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAmountPencePound(t *testing.T) {
	assert.Equal(t, "£0", AmountPence(0).String())
	assert.Equal(t, "£0.01", AmountPence(1).String())
	assert.Equal(t, "£15", AmountPence(1500).String())
	assert.Equal(t, "£103.27", AmountPence(10327).String())
	assert.Equal(t, "£945678.99", AmountPence(94567899).String())
}

func TestAmountPenceInt(t *testing.T) {
	assert.Equal(t, 123, AmountPence(123).Pence())
}
