package sharecodedata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// A Link provides the details of the LPA that will be accessed by a share code.
type Link struct {
	PK        dynamo.ShareKeyType
	SK        dynamo.ShareSortKeyType
	UpdatedAt time.Time

	// LpaKey is the key for the LPA that will be accessed
	LpaKey dynamo.LpaKeyType
	// LpaOwnerKey is the key for the owner of the LPA that will be accessed
	LpaOwnerKey dynamo.LpaOwnerKeyType
	// ActorUID is the UID of the actor being given access to the LPA
	ActorUID actoruid.UID
	// IsReplacementAttorney is true when the actor being given access is being
	// appointed as a replacement (attorney or trust corporation)
	IsReplacementAttorney bool
	// IsTrustCorporation is true when the actor being given access is a trust
	// corporation
	IsTrustCorporation bool

	// The following fields are only relevant to Links for sharing an LPA from a
	// supporter to a donor.

	// InviteSentTo is the email address the supporter sent the invite to
	InviteSentTo string
	// LpaLinkedAt is the time the donor entered the access code
	LpaLinkedAt time.Time
	// LpaLinkedTo is set to the email address the donor used to sign-in when
	// using the code
	LpaLinkedTo string
}
