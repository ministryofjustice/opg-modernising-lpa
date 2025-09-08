package supporterdata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// An LpaLink records data when a donor accesses an LPA created by a supporter.
type LpaLink struct {
	PK        dynamo.LpaKeyType
	SK        dynamo.OrganisationLinkKeyType
	InviteKey dynamo.DonorInviteKeyType

	InviteSentTo string
	InviteSentAt time.Time
	LpaLinkedTo  string
	LpaLinkedAt  time.Time
}
