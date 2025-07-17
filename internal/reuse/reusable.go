package reuse

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

func putReusable(ctx context.Context, dynamoClient DynamoClient, actorType actor.Type, uid actoruid.UID, item any) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID != "" {
		return nil
	}

	if data.SessionID == "" {
		return errors.New("putReusable requires SessionID")
	}

	value, err := attributevalue.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal certificate provider: %w", err)
	}

	return dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actorType.String()), dynamo.MetadataKey(""),
		map[string]string{"#ActorUID": uid.String()},
		map[string]types.AttributeValue{":Value": value},
		"SET #ActorUID = :Value",
	)
}

func reusables[T comparable](ctx context.Context, dynamoClient DynamoClient, actorType actor.Type, seen map[T]struct{}) ([]T, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("reusables requires SessionID")
	}

	var v map[string]orString[T]
	if err := dynamoClient.One(ctx, dynamo.ReuseKey(data.SessionID, actorType.String()), dynamo.MetadataKey(""), &v); err != nil {
		return nil, err
	}

	delete(v, "PK")
	delete(v, "SK")

	var list []T

	if seen == nil {
		seen = map[T]struct{}{}
	}

	for _, item := range v {
		if _, ok := seen[item.v]; !ok {
			list = append(list, item.v)
			seen[item.v] = struct{}{}
		}
	}

	return list, nil
}

func deleteReusable(ctx context.Context, dynamoClient DynamoClient, actorType actor.Type, uid actoruid.UID) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("deleteReusable requires SessionID")
	}

	return dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actorType.String()), dynamo.MetadataKey(""),
		map[string]string{"#ActorUID": uid.String()},
		nil,
		"REMOVE #ActorUID",
	)
}
