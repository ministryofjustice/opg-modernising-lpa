package scheduled

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
)

type DynamoClient interface {
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK, v interface{}) error
	AnyByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	Move(ctx context.Context, oldKeys dynamo.Keys, value any) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type Store struct {
	dynamoClient DynamoClient
	uuidString   func() string
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
		uuidString:   random.UuidString,
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

func (s *Store) Create(ctx context.Context, rows ...Event) error {
	transaction := dynamo.NewTransaction()

	for _, row := range rows {
		row.PK = dynamo.ScheduledDayKey(row.At)
		row.SK = dynamo.ScheduledKey(row.At, s.uuidString())
		row.CreatedAt = s.now()

		transaction.Put(row)
	}

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}

func (s *Store) DeleteAllByUID(ctx context.Context, uid string) error {
	var events []Event

	if err := s.dynamoClient.AllByLpaUIDAndPartialSK(ctx, uid, dynamo.PartialScheduledKey(), &events); err != nil {
		return err
	}

	if len(events) == 0 {
		return fmt.Errorf("no scheduled events found for UID %s", uid)
	}

	var keys []dynamo.Keys
	for _, e := range events {
		keys = append(keys, dynamo.Keys{PK: e.PK, SK: e.SK})
	}

	return s.dynamoClient.DeleteKeys(ctx, keys)
}

func (s *Store) DeleteAllActionByUID(ctx context.Context, actions []Action, uid string) error {
	var events []Event

	if err := s.dynamoClient.AllByLpaUIDAndPartialSK(ctx, uid, dynamo.PartialScheduledKey(), &events); err != nil {
		return err
	}

	var keys []dynamo.Keys
	for _, e := range events {
		if slices.Contains(actions, e.Action) {
			keys = append(keys, dynamo.Keys{PK: e.PK, SK: e.SK})
		}
	}

	return s.dynamoClient.DeleteKeys(ctx, keys)
}
