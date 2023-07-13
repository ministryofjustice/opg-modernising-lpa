package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDataStore) ExpectGet(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("Get", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDataStore) ExpectGetOneByPartialSk(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("GetOneByPartialSk", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDataStore) ExpectGetAllByGsi(ctx, gsi, sk, data interface{}, err error) {
	m.
		On("GetAllByGsi", ctx, gsi, sk, mock.Anything).
		Return(func(ctx context.Context, gsi, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDataStore) ExpectGetAllByKeys(ctx context.Context, keys []dynamo.Key, data interface{}, err error) {
	m.
		On("GetAllByKeys", ctx, keys, mock.Anything).
		Return(func(ctx context.Context, keys []dynamo.Key, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestDonorStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "10100000"}

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetAllByGsi(ctx, "ActorIndex", "#DONOR#an-id",
		[]any{lpa}, nil)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	result, err := donorStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*page.Lpa{lpa}, result)
}

func TestDonorStoreGetAllWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAll(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetAllWhenMissingSessionID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAll(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, nil)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetAnyWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, nil)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestDonorStoreGetWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	ctx := context.Background()
	lpa := &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, lpa).Return(nil)

	donorStore := &donorStore{dataStore: dataStore}

	err := donorStore.Put(ctx, lpa)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenError(t *testing.T) {
	ctx := context.Background()
	lpa := &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, lpa).Return(expectedError)

	donorStore := &donorStore{dataStore: dataStore}

	err := donorStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()
	lpa := &page.Lpa{PK: "LPA#10100000", SK: "#DONOR#an-id", ID: "10100000", UpdatedAt: now, FeeType: page.Full}

	dataStore := newMockDataStore(t)
	dataStore.
		On("Create", ctx, lpa).
		Return(nil)
	dataStore.
		On("Create", ctx, lpaLink{PK: "LPA#10100000", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor}).
		Return(nil)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	result, err := donorStore.Create(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpa, result)
}

func TestDonorStoreCreateWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }, now: func() time.Time { return time.Now() }}

	_, err := donorStore.Create(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreCreateWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()

	testcases := map[string]func(*testing.T) *mockDataStore{
		"certificate provider record": func(t *testing.T) *mockDataStore {
			dataStore := newMockDataStore(t)
			dataStore.
				On("Create", ctx, mock.Anything).
				Return(expectedError)

			return dataStore
		},
		"link record": func(t *testing.T) *mockDataStore {
			dataStore := newMockDataStore(t)
			dataStore.
				On("Create", ctx, mock.Anything).
				Return(nil).
				Once()
			dataStore.
				On("Create", ctx, mock.Anything).
				Return(expectedError)

			return dataStore
		},
	}

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dataStore := makeMockDataStore(t)

			donorStore := &donorStore{
				dataStore:  dataStore,
				uuidString: func() string { return "10100000" },
				now:        func() time.Time { return now },
			}

			_, err := donorStore.Create(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}
