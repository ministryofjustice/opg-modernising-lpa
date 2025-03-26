package dashboarddata

import (
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

// An LpaLink is used to join an actor to an LPA.
type LpaLink struct {
	// PK is the same as the PK for the LPA
	PK dynamo.LpaKeyType
	// SK is the subKey for the current user
	SK dynamo.SubKeyType
	// LpaUID is same as the LpaUID for the LPA
	LpaUID string
	// DonorKey is the donorKey for the donor
	DonorKey dynamo.LpaOwnerKeyType
	// UID is the UID for the linked actor
	UID actoruid.UID
	// ActorType is the type for the current user
	ActorType actor.Type
	// UpdatedAt is set to allow this data to be queried from SKUpdatedAtIndex
	UpdatedAt time.Time
}

func (l LpaLink) UserSub() string {
	if l.SK == "" {
		return ""
	}

	return strings.Split(l.SK.SK(), dynamo.SubKey("").SK())[1]
}
