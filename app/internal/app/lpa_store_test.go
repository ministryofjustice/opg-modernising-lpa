package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDataStore) ExpectGet(ctx, pk, sk, data interface{}, err error) {
	m.
		On("Get", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDataStore) ExpectGetAll(ctx, pk, data interface{}, err error) {
	m.
		On("GetAll", ctx, pk, mock.Anything).
		Return(func(ctx context.Context, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestLpaStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	lpas := []*page.Lpa{{ID: "10100000"}}

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetAll(ctx, "an-id",
		lpas, nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	result, err := lpaStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpas, result)
}

func TestLpaStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "an-id", "123",
		&page.Lpa{ID: "10100000"}, nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "10100000"}, lpa)
}

func TestLpaStoreGetWhenExists(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})
	existingLpa := &page.Lpa{ID: "an-id"}

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "an-id", "123",
		existingLpa, nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, existingLpa, lpa)
}

func TestLpaStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "an-id", "123",
		nil, expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	_, err := lpaStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestLpaStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "an-id", "5", lpa).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}
