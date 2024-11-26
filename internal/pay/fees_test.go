package pay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCost(t *testing.T) {
	testCases := map[string]struct {
		feeType                 FeeType
		previousFee             PreviousFee
		costOfRepeatApplication CostOfRepeatApplication
		expected                int
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
		"previous full": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeFull,
			expected:    4100,
		},
		"previous half": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeHalf,
			expected:    2050,
		},
		"previous exemption": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeExemption,
			expected:    0,
		},
		"previous hardship": {
			feeType:     RepeatApplicationFee,
			previousFee: PreviousFeeHardship,
			expected:    0,
		},
		"repeat entitled to half": {
			feeType:                 RepeatApplicationFee,
			costOfRepeatApplication: CostOfRepeatApplicationHalfFee,
			expected:                4100,
		},
		"repeat entitled to no": {
			feeType:                 RepeatApplicationFee,
			costOfRepeatApplication: CostOfRepeatApplicationNoFee,
			expected:                0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, Cost(tc.feeType, tc.previousFee, tc.costOfRepeatApplication))
		})
	}
}
