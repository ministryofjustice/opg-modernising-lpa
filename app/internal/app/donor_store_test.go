package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
		[]map[string]any{{"Data": lpa}}, nil)

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
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "LPA#5", "#DONOR#an-id", lpa).Return(nil)

	donorStore := &donorStore{dataStore: dataStore}

	err := donorStore.Put(ctx, lpa)
	assert.Nil(t, err)
}

func TestDonorStorePutWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	err := donorStore.Put(ctx, &page.Lpa{})
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStorePutWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "LPA#5", "#DONOR#an-id", lpa).Return(expectedError)

	donorStore := &donorStore{dataStore: dataStore}

	err := donorStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.On("Create", ctx, "LPA#10100000", "#DONOR#an-id", &page.Lpa{ID: "10100000", UpdatedAt: now}).Return(nil)
	dataStore.On("Create", ctx, "LPA#10100000", "#SUB#an-id", "#DONOR#an-id|DONOR").Return(nil)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	lpa, err := donorStore.Create(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "10100000", UpdatedAt: now}, lpa)
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

	dataStore := newMockDataStore(t)
	dataStore.On("Create", ctx, "LPA#10100000", "#DONOR#an-id", &page.Lpa{ID: "10100000", UpdatedAt: now}).Return(expectedError)

	donorStore := &donorStore{dataStore: dataStore, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	_, err := donorStore.Create(ctx)
	assert.Equal(t, expectedError, err)
}
