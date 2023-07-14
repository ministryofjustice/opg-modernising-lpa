package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type attorneyStore struct {
	dataStore DynamoClient
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

	attorney := &actor.AttorneyProvidedDetails{
		PK:            lpaKey(data.LpaID),
		SK:            attorneyKey(data.SessionID),
		ID:            attorneyID,
		LpaID:         data.LpaID,
		UpdatedAt:     s.now(),
		IsReplacement: isReplacement,
	}

	if err := s.dataStore.Create(ctx, attorney); err != nil {
		return nil, err
	}
	if err := s.dataStore.Create(ctx, lpaLink{
		PK:        lpaKey(data.LpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(donorSessionID),
		ActorType: actor.TypeAttorney,
	}); err != nil {
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
	err = s.dataStore.GetAllByGsi(ctx, "ActorIndex", attorneyKey(data.SessionID), &items)

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

	var attorney actor.AttorneyProvidedDetails
	err = s.dataStore.Get(ctx, lpaKey(data.LpaID), attorneyKey(data.SessionID), &attorney)

	return &attorney, err
}

func (s *attorneyStore) Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error {
	attorney.UpdatedAt = s.now()
	return s.dataStore.Put(ctx, attorney)
}

func attorneyKey(s string) string {
	return "#ATTORNEY#" + s
}
