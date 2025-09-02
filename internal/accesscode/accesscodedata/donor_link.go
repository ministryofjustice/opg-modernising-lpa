package accesscodedata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// A DonorLink provides the details of the LPA that will be accessed by a share
// code. It remains once a donor has gained access.
type DonorLink struct {
	PK        dynamo.AccessKeyType
	SK        dynamo.AccessSortKeyType
	UpdatedAt time.Time

	// LpaKey is the key for the LPA that will be accessed
	LpaKey dynamo.LpaKeyType
	// LpaOwnerKey is the key for the owner of the LPA that will be accessed
	LpaOwnerKey dynamo.LpaOwnerKeyType
	// LpaUID is the UID for the LPA that will be accessed
	LpaUID string `dynamodbav:",omitempty"`
	// ActorUID is the UID of the actor being given access to the LPA
	ActorUID actoruid.UID
	// InviteSentTo is the email address the supporter sent the invite to
	InviteSentTo string
	// LpaLinkedAt is the time the donor entered the access code
	LpaLinkedAt time.Time
	// LpaLinkedTo is set to the email address the donor used to sign-in when
	// using the code
	LpaLinkedTo string
}

func (l DonorLink) HasExpired(now time.Time) bool {
	return l.UpdatedAt.AddDate(0, 3, 0).Before(now)
}
