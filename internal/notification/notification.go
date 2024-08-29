package notification

import (
	"time"
)

type Notifications struct {
	FeeEvidence Notification
}

type Notification struct {
	Received time.Time
}
