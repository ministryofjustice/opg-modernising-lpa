package app

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
)

type DynamoUpdateClient interface {
	UpdateReturn(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]types.AttributeValue, expression string) (map[string]types.AttributeValue, error)
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

func (s *uidStore) Set(ctx context.Context, lpaID, sessionID, organisationID, uid string) error {
	values, err := attributevalue.MarshalMap(map[string]any{
		":uid": uid,
		":now": s.now(),
	})
	if err != nil {
		return err
	}

	var sk dynamo.SK = dynamo.DonorKey(sessionID)
	if organisationID != "" {
		sk = dynamo.OrganisationKey(organisationID)
	}

	newAttrs, err := s.dynamoClient.UpdateReturn(ctx, dynamo.LpaKey(lpaID), sk, values,
		"set LpaUID = :uid, UpdatedAt = :now")
	if err != nil {
		return fmt.Errorf("uidStore update failed: %w", err)
	}

	var donor *donordata.Provided
	if err := attributevalue.UnmarshalMap(newAttrs, &donor); err != nil {
		return fmt.Errorf("uidStore unmarshal failed: %w", err)
	}

	if err := s.searchClient.Index(ctx, search.Lpa{
		PK: dynamo.LpaKey(lpaID).PK(),
		SK: sk.SK(),
		Donor: search.LpaDonor{
			FirstNames: donor.Donor.FirstNames,
			LastName:   donor.Donor.LastName,
		},
	}); err != nil {
		return fmt.Errorf("uidStore index failed: %w", err)
	}

	return nil
}
