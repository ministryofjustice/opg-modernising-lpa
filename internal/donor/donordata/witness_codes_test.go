package donordata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWitnessCodeHasExpired(t *testing.T) {
	now := time.Now()

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
			expected: true,
		},
		"15m01s ago": {
			duration: 15*time.Minute + time.Second,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			codes := WitnessCodes{
				{Code: "a", Created: now.Add(-tc.duration)},
			}

			code, _ := codes.Find("a")
			assert.Equal(t, tc.expected, code.HasExpired())
		})
	}
}

func TestWitnessCodesFind(t *testing.T) {
	codes := WitnessCodes{
		{Code: "new", Created: time.Now()},
		{Code: "expired", Created: time.Now().Add(-16 * time.Minute)},
		{Code: "almost ignored", Created: time.Now().Add(-2*time.Hour + time.Second)},
		{Code: "ignored", Created: time.Now().Add(-2 * time.Hour)},
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
			_, ok := codes.Find(code)
			assert.Equal(t, expected, ok)
		})
	}
}

func TestWitnessCodesCanRequest(t *testing.T) {
	now := time.Now()

	testcases := map[string]struct {
		codes    WitnessCodes
		expected bool
	}{
		"empty": {
			expected: true,
		},
		"after 1 minute": {
			codes:    WitnessCodes{{Created: now.Add(-time.Minute - time.Second)}},
			expected: true,
		},
		"within 1 minute": {
			codes:    WitnessCodes{{Created: now.Add(-time.Minute)}},
			expected: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.codes.CanRequest(now))
		})
	}
}
