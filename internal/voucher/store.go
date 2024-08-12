package voucher

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

type Store struct {
	dynamoClient any
}

func NewStore(dynamoClient any) *Store {
	return &Store{dynamoClient: dynamoClient}
}

func (s *Store) Create(ctx context.Context, shareCode sharecodedata.Link, email string) (any, error) {
	return nil, nil
}
