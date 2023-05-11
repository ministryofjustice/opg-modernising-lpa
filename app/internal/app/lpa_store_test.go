package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDataStore) ExpectGet(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("GetOneByPartialSk", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDataStore) ExpectGetAll(ctx, gsi, sk, data interface{}, err error) {
	m.
		On("GetAllByGsi", ctx, gsi, sk, mock.Anything).
		Return(func(ctx context.Context, gsi, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestLpaStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	lpas := []*page.Lpa{{ID: "10100000"}}

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetAll(ctx, "ActorIndex", "#DONOR#an-id", lpas, nil)

	lpaStore := &lpaStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	result, err := lpaStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpas, result)
}

func TestLpaStoreGetAllWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	lpaStore := &lpaStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := lpaStore.GetAll(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestLpaStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, nil)

	lpaStore := &lpaStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestLpaStoreGetWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	lpaStore := &lpaStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	_, err := lpaStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestLpaStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, uuidString: func() string { return "10100000" }}

	_, err := lpaStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestLpaStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "LPA#5", "#DONOR#an-id", lpa).Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, lpa)
	assert.Nil(t, err)
}

func TestLpaStorePutWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	lpaStore := &lpaStore{dataStore: nil, uuidString: func() string { return "10100000" }}

	err := lpaStore.Put(ctx, &page.Lpa{})
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestLpaStorePutWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "LPA#5", "#DONOR#an-id", lpa).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}

func TestLpaStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.On("Create", ctx, "LPA#10100000", "#DONOR#an-id", &page.Lpa{ID: "10100000", UpdatedAt: now}).Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	lpa, err := lpaStore.Create(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "10100000", UpdatedAt: now}, lpa)
}

func TestLpaStoreCreateWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	lpaStore := &lpaStore{dataStore: nil, uuidString: func() string { return "10100000" }, now: func() time.Time { return time.Now() }}

	_, err := lpaStore.Create(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestLpaStoreCreateWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.On("Create", ctx, "LPA#10100000", "#DONOR#an-id", &page.Lpa{ID: "10100000", UpdatedAt: now}).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	_, err := lpaStore.Create(ctx)
	assert.Equal(t, expectedError, err)
}
