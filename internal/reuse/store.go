package reuse

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, names map[string]string, values map[string]types.AttributeValue, expression string) error
}

type Store struct {
	dynamoClient DynamoClient
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
	}
}

func (s *Store) PutCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("reuseStore.AddCorrespondent requires SessionID")
	}

	actorUID := correspondent.UID
	correspondent.UID = actoruid.UID{}
	value, err := attributevalue.Marshal(correspondent)
	if err != nil {
		return fmt.Errorf("marshal correspondent: %w", err)
	}

	return s.dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actor.TypeCorrespondent.String()), dynamo.MetadataKey(""),
		map[string]string{"#ActorUID": actorUID.String()},
		map[string]types.AttributeValue{":Value": value},
		"SET #ActorUID = :Value",
	)
}

func (s *Store) DeleteCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("reuseStore.DeleteCorrespondent requires SessionID")
	}

	return s.dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actor.TypeCorrespondent.String()), dynamo.MetadataKey(""),
		map[string]string{"#ActorUID": correspondent.UID.String()},
		nil,
		"REMOVE #ActorUID",
	)
}

func (s *Store) Correspondents(ctx context.Context) ([]donordata.Correspondent, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("reuseStore.Correspondents requires SessionID")
	}

	var v map[string]orString[donordata.Correspondent]
	if err := s.dynamoClient.One(ctx, dynamo.ReuseKey(data.SessionID, actor.TypeCorrespondent.String()), dynamo.MetadataKey(""), &v); err != nil {
		return nil, err
	}

	delete(v, "PK")
	delete(v, "SK")

	var correspondents []donordata.Correspondent
	seen := map[donordata.Correspondent]struct{}{}

	for _, correspondent := range v {
		if _, ok := seen[correspondent.v]; !ok {
			correspondents = append(correspondents, correspondent.v)
			seen[correspondent.v] = struct{}{}
		}
	}

	slices.SortFunc(correspondents, func(a, b donordata.Correspondent) int {
		return strings.Compare(a.FullName(), b.FullName())
	})

	return correspondents, nil
}
