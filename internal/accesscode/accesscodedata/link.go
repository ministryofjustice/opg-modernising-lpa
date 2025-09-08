package accesscodedata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// A Link provides the details of the LPA that will be accessed by a share code.
type Link struct {
	PK        dynamo.AccessKeyType
	SK        dynamo.AccessSortKeyType
	UpdatedAt time.Time
	ExpiresAt time.Time

	// LpaKey is the key for the LPA that will be accessed
	LpaKey dynamo.LpaKeyType
	// LpaOwnerKey is the key for the owner of the LPA that will be accessed
	LpaOwnerKey dynamo.LpaOwnerKeyType
	// LpaUID is the UID for the LPA that will be accessed
	LpaUID string `dynamodbav:",omitempty"`
	// ActorUID is the UID of the actor being given access to the LPA
	ActorUID actoruid.UID
	// IsReplacementAttorney is true when the actor being given access is being
	// appointed as a replacement (attorney or trust corporation)
	IsReplacementAttorney bool
	// IsTrustCorporation is true when the actor being given access is a trust
	// corporation
	IsTrustCorporation bool
}

// For must be used to get the Link to store.
func (l Link) For(now time.Time) Link {
	l.UpdatedAt = now
	if l.PK.IsDonor() {
		l.ExpiresAt = now.AddDate(0, 3, 0)
	} else {
		l.ExpiresAt = now.AddDate(2, 0, 0)
	}
	return l
}
