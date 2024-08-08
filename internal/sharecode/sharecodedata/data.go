package sharecodedata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type Data struct {
	PK                    dynamo.ShareKeyType
	SK                    dynamo.ShareSortKeyType
	UpdatedAt             time.Time
	LpaKey                dynamo.LpaKeyType
	LpaOwnerKey           dynamo.LpaOwnerKeyType
	ActorUID              actoruid.UID
	IsReplacementAttorney bool
	IsTrustCorporation    bool

	// InviteSentTo is the email address the supporter sent the invite to
	InviteSentTo string
	// LpaLinkedAt is the time the donor entered the access code
	LpaLinkedAt time.Time
	// LpaLinkedTo is set to the email address the donor used to sign-in when
	// using the code
	LpaLinkedTo string
}
