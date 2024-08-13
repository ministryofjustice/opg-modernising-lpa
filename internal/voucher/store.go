package voucher

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type Store struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{dynamoClient: dynamoClient, now: time.Now}
}

func (s *Store) Create(ctx context.Context, shareCode sharecodedata.Link, email string) (*voucherdata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("voucher.Store.Create requires LpaID and SessionID")
	}

	provided := &voucherdata.Provided{
		PK:        dynamo.LpaKey(data.LpaID),
		SK:        dynamo.VoucherKey(data.SessionID),
		LpaID:     data.LpaID,
		UpdatedAt: s.now(),
		Email:     email,
	}

	transaction := dynamo.NewTransaction().
		Create(provided).
		Create(dashboarddata.LpaLink{
			PK:        dynamo.LpaKey(data.LpaID),
			SK:        dynamo.SubKey(data.SessionID),
			DonorKey:  shareCode.LpaOwnerKey,
			ActorType: actor.TypeVoucher,
			UpdatedAt: s.now(),
		}).
		Delete(dynamo.Keys{PK: shareCode.PK, SK: shareCode.SK})

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return provided, err
}

func (s *Store) Get(ctx context.Context) (*voucherdata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("voucher.Store.Get requires LpaID and SessionID")
	}

	var provided voucherdata.Provided
	err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.VoucherKey(data.SessionID), &provided)

	return &provided, err
}
