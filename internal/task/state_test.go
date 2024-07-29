package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskStateString(t *testing.T) {
	testCases := []struct {
		State    State
		Expected string
	}{
		{
			State:    StateNotStarted,
			Expected: "notStarted",
		},
		{
			State:    StateInProgress,
			Expected: "inProgress",
		},
		{
			State:    StateCompleted,
			Expected: "completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Expected, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.State.String())
		})
	}
}
