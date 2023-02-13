package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockDataStore struct {
	data interface{}
	mock.Mock
}

func (m *mockDataStore) GetAll(ctx context.Context, pk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk).Error(0)
}

func (m *mockDataStore) Get(ctx context.Context, pk, sk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk, sk).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, pk, sk string, v interface{}) error {
	return m.Called(ctx, pk, sk, v).Error(0)
}

func TestLpaStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	lpas := []*page.Lpa{{ID: "10100000"}}

	dataStore := &mockDataStore{data: lpas}
	dataStore.On("GetAll", ctx, "an-id").Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	result, err := lpaStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpas, result)
}

func TestLpaStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dataStore := &mockDataStore{data: &page.Lpa{ID: "10100000"}}
	dataStore.On("Get", ctx, "an-id", "123").Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "10100000"}, lpa)
}

func TestLpaStoreGetWhenExists(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})
	existingLpa := &page.Lpa{ID: "an-id"}

	dataStore := &mockDataStore{data: existingLpa}
	dataStore.On("Get", ctx, "an-id", "123").Return(nil)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	lpa, err := lpaStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, existingLpa, lpa)
}

func TestLpaStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dataStore := &mockDataStore{}
	dataStore.On("Get", ctx, "an-id", "123").Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	_, err := lpaStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestLpaStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})
	lpa := &page.Lpa{ID: "5"}

	dataStore := &mockDataStore{}
	dataStore.On("Put", ctx, "an-id", "5", lpa).Return(expectedError)

	lpaStore := &lpaStore{dataStore: dataStore}

	err := lpaStore.Put(ctx, lpa)
	assert.Equal(t, expectedError, err)
}
