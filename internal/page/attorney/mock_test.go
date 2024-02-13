package attorney

import (
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

var (
	testUID       = actoruid.New()
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeAttorney,
	}
	testReplacementAppData = page.AppData{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeReplacementAttorney,
	}
	testTrustCorporationAppData = page.AppData{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeTrustCorporation,
	}
	testReplacementTrustCorporationAppData = page.AppData{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeReplacementTrustCorporation,
	}
)
