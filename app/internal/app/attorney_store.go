package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type attorneyStore struct {
	dataStore DataStore
	now       func() time.Time
}

func (s *attorneyStore) Create(ctx context.Context, donorSessionID, attorneyID string, isReplacement bool) (*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Create requires LpaID and SessionID")
	}

	attorney := &actor.AttorneyProvidedDetails{ID: attorneyID, LpaID: data.LpaID, UpdatedAt: s.now(), IsReplacement: isReplacement}

	pk, sk, subk := makeAttorneyKeys(data.LpaID, data.SessionID)
	if err := s.dataStore.Create(ctx, pk, sk, attorney); err != nil {
		return nil, err
	}
	if err := s.dataStore.Create(ctx, pk, subk, "#DONOR#"+donorSessionID+"|ATTORNEY"); err != nil {
		return nil, err
	}

	return attorney, err
}

func (s *attorneyStore) GetAll(ctx context.Context) ([]*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("attorneyStore.GetAll requires SessionID")
	}

	var items []*actor.AttorneyProvidedDetails
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", "#ATTORNEY#"+data.SessionID, &items)

	return items, err
}

func (s *attorneyStore) Get(ctx context.Context) (*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Get requires LpaID and SessionID")
	}

	pk, sk, _ := makeAttorneyKeys(data.LpaID, data.SessionID)

	var attorney actor.AttorneyProvidedDetails
	err = s.dataStore.Get(ctx, pk, sk, &attorney)

	return &attorney, err
}

func (s *attorneyStore) Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return errors.New("attorneyStore.Put requires LpaID and SessionID")
	}

	pk, sk, _ := makeAttorneyKeys(data.LpaID, data.SessionID)

	attorney.UpdatedAt = s.now()
	return s.dataStore.Put(ctx, pk, sk, attorney)
}

func makeAttorneyKeys(lpaID, sessionID string) (string, string, string) {
	return "LPA#" + lpaID, "#ATTORNEY#" + sessionID, "#SUB#" + sessionID
}
