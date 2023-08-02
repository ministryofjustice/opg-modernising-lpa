package attorney

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

var (
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID:  "session-id",
		LpaID:      "lpa-id",
		AttorneyID: "attorney-id",
		Lang:       localize.En,
		Paths:      page.Paths,
		ActorType:  actor.TypeAttorney,
	}
	testReplacementAppData = page.AppData{
		SessionID:  "session-id",
		LpaID:      "lpa-id",
		AttorneyID: "attorney-id",
		Lang:       localize.En,
		Paths:      page.Paths,
		ActorType:  actor.TypeReplacementAttorney,
	}
	testTrustCorporationAppData = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
		ActorType: actor.TypeAttorney,
	}
)
