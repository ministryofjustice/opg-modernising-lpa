package cron

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// A Row specifies an action to take in the future.
type Row struct {
	PK dynamo.CronDayKeyType
	SK dynamo.CronKeyType
	// CreatedAt is when the Row was created
	CreatedAt time.Time
	// LockedUntil is set to show the Row is running, it cannot be retrieved until this time has passed
	LockedUntil time.Time
	// HandledAt is set when the Row has finished running
	HandledAt time.Time
	// At is when the action should be done
	At time.Time
	// Action is what to do when run
	Action Action
	// TargetPK is the PK of what will be affected by the action
	TargetPK dynamo.PKType
	// TargetSK is the SK of what will be affected by the action
	TargetSK dynamo.SKType
}
