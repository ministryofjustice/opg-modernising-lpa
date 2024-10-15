package scheduled

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	Move(ctx context.Context, oldKeys dynamo.Keys, value any) error
	AnyByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	Put(ctx context.Context, v interface{}) error
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
		now:          time.Now,
	}
}

func (s *Store) Pop(ctx context.Context, day time.Time) (*Event, error) {
	var row Event
	if err := s.dynamoClient.AnyByPK(ctx, dynamo.ScheduledDayKey(day), &row); err != nil {
		return nil, err
	}

	oldKeys := dynamo.Keys{PK: row.PK, SK: row.SK}
	row.PK = row.PK.Handled()

	if err := s.dynamoClient.Move(ctx, oldKeys, row); err != nil {
		return nil, err
	}

	return &row, nil
}

func (s *Store) Put(ctx context.Context, row Event) error {
	row.PK = dynamo.ScheduledDayKey(row.At)
	row.SK = dynamo.ScheduledKey(row.At, int(row.Action))
	row.CreatedAt = s.now()

	return s.dynamoClient.Put(ctx, row)
}
