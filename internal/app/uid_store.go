package app

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
)

type DynamoUpdateClient interface {
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type SearchClient interface {
	Index(ctx context.Context, lpa search.Lpa) error
}

type uidStore struct {
	dynamoClient DynamoUpdateClient
	now          func() time.Time
	searchClient SearchClient
}

func NewUidStore(dynamoClient DynamoUpdateClient, searchClient SearchClient, now func() time.Time) *uidStore {
	return &uidStore{dynamoClient: dynamoClient, searchClient: searchClient, now: now}
}

func (s *uidStore) Set(ctx context.Context, provided *donordata.Provided, uid string) error {
	provided.LpaUID = uid
	provided.UpdatedAt = s.now()

	if err := s.searchClient.Index(ctx, search.Lpa{
		PK: provided.PK.PK(),
		SK: provided.SK.SK(),
		Donor: search.LpaDonor{
			FirstNames: provided.Donor.FirstNames,
			LastName:   provided.Donor.LastName,
		},
	}); err != nil {
		return fmt.Errorf("uidStore index failed: %w", err)
	}

	transaction := dynamo.NewTransaction().
		Put(provided).
		Create(dynamo.Keys{PK: dynamo.UIDKey(uid), SK: dynamo.MetadataKey("")})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return fmt.Errorf("uidStore update failed: %w", err)
	}

	return nil
}
