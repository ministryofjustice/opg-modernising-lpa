package scheduled

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// A Event specifies an action to take in the future.
type Event struct {
	PK dynamo.ScheduledDayKeyType
	SK dynamo.ScheduledKeyType
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
	// TargetLpaKey is used to specify the target of the action
	TargetLpaKey dynamo.LpaKeyType
	// TargetLpaOwnerKey is used to specify the target of the action
	TargetLpaOwnerKey dynamo.LpaOwnerKeyType
}
