package attorneypage

import (
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

var (
	testUID       = actoruid.New()
	expectedError = errors.New("err")
	testAppData   = appcontext.Data{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeAttorney,
	}
	testReplacementAppData = appcontext.Data{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeReplacementAttorney,
	}
	testTrustCorporationAppData = appcontext.Data{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeTrustCorporation,
	}
	testReplacementTrustCorporationAppData = appcontext.Data{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		AttorneyUID: testUID,
		Lang:        localize.En,
		ActorType:   actor.TypeReplacementTrustCorporation,
	}
)

func evalT[T any](fn func(*testing.T) T, t *testing.T) T {
	if fn == nil {
		return *new(T)
	}

	return fn(t)
}
