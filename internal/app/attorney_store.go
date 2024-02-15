package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type attorneyStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *attorneyStore) Create(ctx context.Context, donorSessionID string, attorneyUID actoruid.UID, isReplacement, isTrustCorporation bool) (*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Create requires LpaID and SessionID")
	}

	attorney := &actor.AttorneyProvidedDetails{
		PK:                 lpaKey(data.LpaID),
		SK:                 attorneyKey(data.SessionID),
		UID:                attorneyUID,
		LpaID:              data.LpaID,
		UpdatedAt:          s.now(),
		IsReplacement:      isReplacement,
		IsTrustCorporation: isTrustCorporation,
	}

	if err := s.dynamoClient.Create(ctx, attorney); err != nil {
		return nil, err
	}
	if err := s.dynamoClient.Create(ctx, lpaLink{
		PK:        lpaKey(data.LpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(donorSessionID),
		ActorType: actor.TypeAttorney,
		UpdatedAt: s.now(),
	}); err != nil {
		return nil, err
	}

	return attorney, err
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
	err = s.dynamoClient.One(ctx, lpaKey(data.LpaID), attorneyKey(data.SessionID), &attorney)

	return &attorney, err
}

func (s *attorneyStore) GetAny(ctx context.Context) ([]*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("attorneyStore.GetAny requires LpaID")
	}

	var attorneys []*actor.AttorneyProvidedDetails
	err = s.dynamoClient.AllByPartialSK(ctx, lpaKey(data.LpaID), "#ATTORNEY#", &attorneys)

	return attorneys, err
}

func (s *attorneyStore) Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error {
	attorney.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, attorney)
}

func attorneyKey(s string) string {
	return "#ATTORNEY#" + s
}
