package scheduled

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
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

func (s *Store) Pop(ctx context.Context, day time.Time) (Event, error) {
	var row Event
	if err := s.dynamoClient.OneByPK(ctx, dynamo.ScheduledDayKey(day), &row); err != nil {
		return row, err
	}

	// TODO: need to check that this ensures an item was deleted, otherwise we
	// might have 2 runners running the same item
	if err := s.dynamoClient.DeleteOne(ctx, row.PK, row.SK); err != nil {
		return row, err
	}

	// TODO: A better approach may be to do a transaction where we delete the item
	// and put a copy of it somewhere else (like a DLQ). Then later if successful
	// we can remove from the DLQ, and if not it can always be manually dealt
	// with?

	return row, nil
}

func (s *Store) Put(ctx context.Context, row Event) error {
	row.PK = dynamo.ScheduledDayKey(row.At)
	row.SK = dynamo.ScheduledKey(row.At, int(row.Action))
	row.CreatedAt = s.now()

	return s.dynamoClient.Put(ctx, row)
}
