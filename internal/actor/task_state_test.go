package actor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskStateString(t *testing.T) {
	testCases := []struct {
		State    TaskState
		Expected string
	}{
		{
			State:    TaskNotStarted,
			Expected: "notStarted",
		},
		{
			State:    TaskInProgress,
			Expected: "inProgress",
		},
		{
			State:    TaskCompleted,
			Expected: "completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Expected, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.State.String())
		})
	}
}
