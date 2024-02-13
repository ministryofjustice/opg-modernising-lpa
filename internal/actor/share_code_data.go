package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"

type ShareCodeData struct {
	PK, SK                string
	SessionID             string
	LpaID                 string
	ActorUID              actoruid.UID
	IsReplacementAttorney bool
	IsTrustCorporation    bool
	DonorFullname         string
	DonorFirstNames       string
}
