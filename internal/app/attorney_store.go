package app

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type attorneyStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *attorneyStore) Create(ctx context.Context, shareCode actor.ShareCodeData, email string) (*actor.AttorneyProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("attorneyStore.Create requires LpaID and SessionID")
	}

	attorney := &actor.AttorneyProvidedDetails{
		PK:                 dynamo.LpaKey(data.LpaID),
		SK:                 dynamo.AttorneyKey(data.SessionID),
		UID:                shareCode.ActorUID,
		LpaID:              data.LpaID,
		UpdatedAt:          s.now(),
		IsReplacement:      shareCode.IsReplacementAttorney,
		IsTrustCorporation: shareCode.IsTrustCorporation,
		Email:              email,
	}

	transaction := dynamo.NewTransaction().
		Create(attorney).
		Create(lpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			DonorKey:  shareCode.LpaOwnerKey,
			ActorType: actor.TypeAttorney,
			UpdatedAt: s.now(),
		}).
		Delete(dynamo.Keys{PK: shareCode.PK, SK: shareCode.SK})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
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
	err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.AttorneyKey(data.SessionID), &attorney)

	return &attorney, err
}

func (s *attorneyStore) Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error {
	attorney.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, attorney)
}
