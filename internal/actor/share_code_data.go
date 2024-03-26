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
	DonorActingOn         ActingOn

	// InviteSentTo is the email address the supporter sent the invite to
	InviteSentTo string
	// LpaLinkedAt is the time the donor entered the access code
	LpaLinkedAt time.Time
	// LpaLinkedTo is set to the email address the donor used to sign-in when
	// using the code
	LpaLinkedTo string
}
