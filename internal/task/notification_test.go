package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationSubjectsHasSeen(t *testing.T) {
	notifications := NotificationSubjects{NotificationSuccessfulVouch}

	assert.True(t, notifications.HasSeen(NotificationSuccessfulVouch))
	assert.False(t, notifications.HasSeen(NotificationSubject(99)))
}
