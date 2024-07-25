package attorneypage

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type Localizer interface {
	page.Localizer
}

type Template func(io.Writer, interface{}) error

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, details *actor.AttorneyProvidedDetails) error

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpastore.Lpa, error)
}

type AttorneyStore interface {
	Create(ctx context.Context, shareCode actor.ShareCodeData, email string) (*actor.AttorneyProvidedDetails, error)
	Get(ctx context.Context) (*actor.AttorneyProvidedDetails, error)
	Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error
	Delete(ctx context.Context) error
}

type NotifyClient interface {
	SendActorEmail(ctx context.Context, to, lpaUID string, email notify.Email) error
}

func findAttorneyFullName(lpa *lpastore.Lpa, uid actoruid.UID, isTrustCorporation, isReplacement bool) (string, error) {
	if t := lpa.ReplacementAttorneys.TrustCorporation; t.UID == uid {
		return t.Name, nil
	}

	if t := lpa.Attorneys.TrustCorporation; t.UID == uid {
		return t.Name, nil
	}

	if a, ok := lpa.ReplacementAttorneys.Get(uid); ok {
		return a.FullName(), nil
	}

	if a, ok := lpa.Attorneys.Get(uid); ok {
		return a.FullName(), nil
	}

	return "", errors.New("attorney not found")
}
