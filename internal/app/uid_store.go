package app

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
)

type DynamoUpdateClient interface {
	UpdateReturn(ctx context.Context, pk, sk string, values map[string]types.AttributeValue, expression string) (map[string]types.AttributeValue, error)
}

type SearchClient interface {
	Index(ctx context.Context, lpa search.Lpa) error
}

type uidStore struct {
	dynamoClient        DynamoUpdateClient
	now                 func() time.Time
	searchClientFactory func() (SearchClient, error)
	searchClient        SearchClient
}

func NewUidStore(dynamoClient DynamoUpdateClient, searchClientFactory func() (SearchClient, error), now func() time.Time) *uidStore {
	return &uidStore{dynamoClient: dynamoClient, searchClientFactory: searchClientFactory, now: now}
}

func (s *uidStore) Set(ctx context.Context, lpaID, sessionID, organisationID, uid string) error {
	values, err := attributevalue.MarshalMap(map[string]any{
		":uid": uid,
		":now": s.now(),
	})
	if err != nil {
		return err
	}

	sk := donorKey(sessionID)
	if organisationID != "" {
		sk = organisationKey(organisationID)
	}

	newAttrs, err := s.dynamoClient.UpdateReturn(ctx, lpaKey(lpaID), sk, values,
		"set LpaUID = :uid, UpdatedAt = :now")
	if err != nil {
		return fmt.Errorf("uidStore update failed: %w", err)
	}

	var donor *actor.DonorProvidedDetails
	if err := attributevalue.UnmarshalMap(newAttrs, &donor); err != nil {
		return fmt.Errorf("uidStore unmarshal failed: %w", err)
	}

	if s.searchClient == nil {
		s.searchClient, err = s.searchClientFactory()
		if err != nil {
			return fmt.Errorf("uidStore could not create search client: %w", err)
		}
	}

	if err := s.searchClient.Index(ctx, search.Lpa{
		PK:            lpaKey(lpaID),
		SK:            sk,
		DonorFullName: donor.Donor.FullName(),
	}); err != nil {
		return fmt.Errorf("uidStore index failed: %w", err)
	}

	return nil
}
