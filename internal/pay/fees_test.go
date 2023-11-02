package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCost(t *testing.T) {
	testCases := map[string]struct {
		feeType     FeeType
		previousFee PreviousFee
		expected    int
	}{
		"full": {
			feeType:  FullFee,
			expected: 8200,
		},
		"half": {
			feeType:  HalfFee,
			expected: 4100,
		},
		"no fee": {
			feeType:  NoFee,
			expected: 0,
		},
		"hardship": {
			feeType:  HardshipFee,
			expected: 0,
		},
		"repeat full": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeFull,
			expected:    4100,
		},
		"repeat half": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeHalf,
			expected:    2050,
		},
		"repeat exemption": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeExemption,
			expected:    0,
		},
		"repeat hardship": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeHardship,
			expected:    0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Cost(tc.feeType, tc.previousFee))
		})
	}
}
