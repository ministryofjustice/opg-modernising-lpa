package place

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressEqual(t *testing.T) {
	testcases := map[string]struct {
		a, b     Address
		expected bool
	}{
		"same":      {Address{Line1: "x", Postcode: "y"}, Address{Line1: "x", Postcode: "y"}, true},
		"different": {Address{Line1: "x", Postcode: "y"}, Address{Line1: "y", Postcode: "x"}, false},
		"same caps": {Address{Line1: "x", Postcode: "y"}, Address{Line1: "X", Postcode: "Y"}, true},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.a.Equal(tc.b))
		})
	}
}
