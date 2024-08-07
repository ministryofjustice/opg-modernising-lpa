package donordata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testNow = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)

func TestWitnessCodeHasExpired(t *testing.T) {
	testCases := map[string]struct {
		duration time.Duration
		expected bool
	}{
		"now": {
			duration: 0,
			expected: false,
		},
		"14m59s ago": {
			duration: 14*time.Minute + 59*time.Second,
			expected: false,
		},
		"15m ago": {
			duration: 15 * time.Minute,
			expected: false,
		},
		"15m01s ago": {
			duration: 15*time.Minute + time.Second,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			codes := WitnessCodes{
				{Code: "a", Created: testNow.Add(-tc.duration)},
			}

			code, _ := codes.Find("a", testNow)
			assert.Equal(t, tc.expected, code.HasExpired(testNow))
		})
	}
}

func TestWitnessCodesFind(t *testing.T) {
	codes := WitnessCodes{
		{Code: "new", Created: testNow},
		{Code: "expired", Created: testNow.Add(-16 * time.Minute)},
		{Code: "almost ignored", Created: testNow.Add(-2 * time.Hour)},
		{Code: "ignored", Created: testNow.Add(-2*time.Hour - time.Second)},
	}

	testcases := map[string]bool{
		"wrong":          false,
		"new":            true,
		"expired":        true,
		"almost ignored": true,
		"ignored":        false,
	}

	for code, expected := range testcases {
		t.Run(code, func(t *testing.T) {
			_, ok := codes.Find(code, testNow)
			assert.Equal(t, expected, ok)
		})
	}
}

func TestWitnessCodesCanRequest(t *testing.T) {
	testcases := map[string]struct {
		codes    WitnessCodes
		expected bool
	}{
		"empty": {
			expected: true,
		},
		"after 1 minute": {
			codes:    WitnessCodes{{Created: testNow.Add(-time.Minute - time.Second)}},
			expected: true,
		},
		"within 1 minute": {
			codes:    WitnessCodes{{Created: testNow.Add(-time.Minute)}},
			expected: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.codes.CanRequest(testNow))
		})
	}
}
