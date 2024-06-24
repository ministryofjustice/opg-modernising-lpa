package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVoucherFullName(t *testing.T) {
	assert.Equal(t, "John Smith", Voucher{FirstNames: "John", LastName: "Smith"}.FullName())
}
