package voucher

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), "a", "b")
	expectedError = errors.New("expected")
)

func (m *mockDynamoClient) ExpectOne(ctx, pk, sk, data interface{}, err error) {
	m.
		On("One", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (c *mockDynamoClient_OneByPartialSK_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func TestNewStore(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)

	store := NewStore(dynamoClient)
	assert.Equal(t, dynamoClient, store.dynamoClient)
}

func TestVoucherStoreCreate(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})
	now := time.Now()
	uid := actoruid.New()
	details := &voucherdata.Provided{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.VoucherKey("456"),
		LpaID:     "123",
		UpdatedAt: now,
		Email:     "a@example.com",
	}

	accessCode := accesscodedata.Link{
		PK:          dynamo.AccessKey(dynamo.VoucherAccessKey("123")),
		SK:          dynamo.ShareSortKey(dynamo.MetadataKey("123")),
		ActorUID:    uid,
		UpdatedAt:   now,
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	expectedTransaction := &dynamo.Transaction{
		Creates: []any{
			details,
			dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.ReservedKey(dynamo.VoucherKey)},
			dashboarddata.LpaLink{
				PK:        dynamo.LpaKey("123"),
				SK:        dynamo.SubKey("456"),
				DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				ActorType: actor.TypeVoucher,
				UpdatedAt: now,
			},
		},
		Deletes: []dynamo.Keys{
			{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("123")),
				SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
			},
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, expectedTransaction).
		Return(nil)

	store := Store{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	provided, err := store.Create(ctx, accessCode, "a@example.com")
	assert.Nil(t, err)
	assert.Equal(t, details, provided)
}

func TestVoucherStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	store := &Store{dynamoClient: nil, now: nil}

	_, err := store.Create(ctx, accesscodedata.Link{}, "")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestVoucherStoreCreateWhenSessionMissingRequiredData(t *testing.T) {
	testcases := map[string]*appcontext.Session{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), sessionData)

			store := &Store{}

			_, err := store.Create(ctx, accesscodedata.Link{}, "")
			assert.NotNil(t, err)
		})
	}
}

func TestVoucherStoreCreateWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	_, err := store.Create(ctx, accesscodedata.Link{
		PK: dynamo.AccessKey(dynamo.VoucherAccessKey("123")),
		SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
	}, "")
	assert.Equal(t, expectedError, err)
}

func TestVoucherStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.VoucherKey("456"),
			&voucherdata.Provided{LpaID: "123"}, nil)

	store := &Store{dynamoClient: dynamoClient, now: nil}

	provided, err := store.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &voucherdata.Provided{LpaID: "123"}, provided)
}

func TestVoucherStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	store := &Store{dynamoClient: nil, now: nil}

	_, err := store.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestVoucherStoreGetMissingLpaIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "456"})

	store := &Store{}

	_, err := store.Get(ctx)
	assert.Equal(t, errors.New("voucher.Store.Get requires LpaID and SessionID"), err)
}

func TestVoucherStoreGetMissingSessionIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	store := &Store{}

	_, err := store.Get(ctx)
	assert.Equal(t, errors.New("voucher.Store.Get requires LpaID and SessionID"), err)
}

func TestVoucherStoreGetOnError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.VoucherKey("456"),
			&voucherdata.Provided{LpaID: "123"}, expectedError)

	store := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := store.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestVoucherStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &voucherdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.VoucherKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	store := &Store{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := store.Put(ctx, &voucherdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.VoucherKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestVoucherStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.VoucherKey(""),
			mock.Anything).
		Return(nil).
		SetData(&voucherdata.Provided{LpaID: "123"})

	store := &Store{dynamoClient: dynamoClient, now: nil}

	provided, err := store.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &voucherdata.Provided{LpaID: "123"}, provided)
}

func TestVoucherStoreGetAnyWhenMissingLpaID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	store := &Store{dynamoClient: nil, now: nil}

	_, err := store.GetAny(ctx)
	assert.ErrorContains(t, err, "voucher.Store.GetAny requires LpaID")
}

func TestVoucherStoreGetAnyWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByPartialSK(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := store.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}
