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

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	result, err := lpaStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpas, result)
}

func TestLpaStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestLpaStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGet(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

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

func TestLpaStorePutWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := newMockDataStore(t)
	dataStore.On("Put", ctx, "LPA#5", "#DONOR#an-id", lpa).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}
