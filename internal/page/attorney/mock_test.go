package attorney

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

var (
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID:  "session-id",
		LpaID:      "lpa-id",
		AttorneyID: "attorney-id",
		Lang:       localize.En,
		ActorType:  actor.TypeAttorney,
	}
	testReplacementAppData = page.AppData{
		SessionID:  "session-id",
		LpaID:      "lpa-id",
		AttorneyID: "attorney-id",
		Lang:       localize.En,
		ActorType:  actor.TypeReplacementAttorney,
	}
	testTrustCorporationAppData = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		ActorType: actor.TypeAttorney,
	}
	testReplacementTrustCorporationAppData = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		ActorType: actor.TypeReplacementAttorney,
	}
)
