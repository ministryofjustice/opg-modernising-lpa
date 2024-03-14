package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

type ShareCodeData struct {
	PK, SK                string
	UpdatedAt             time.Time
	SessionID             string
	LpaID                 string
	ActorUID              actoruid.UID
	IsReplacementAttorney bool
	IsTrustCorporation    bool

	// InviteSentTo is the email address the supporter sent the invite to
	InviteSentTo string
}
