package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestAttorneyStoreCreate(t *testing.T) {
	for name, isReplacement := range map[string]bool{"attorney": false, "replacement": true} {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
			now := time.Now()
			details := &actor.AttorneyProvidedDetails{PK: "LPA#123", SK: "#ATTORNEY#456", ID: "attorney-id", LpaID: "123", UpdatedAt: now, IsReplacement: isReplacement}

			dataStore := newMockDataStore(t)
			dataStore.
				On("Create", ctx, details).
				Return(nil)
			dataStore.
				On("Create", ctx, lpaLink{PK: "LPA#123", SK: "#SUB#456", DonorKey: "#DONOR#session-id", ActorType: actor.TypeAttorney}).
				Return(nil)

			attorneyStore := &attorneyStore{dataStore: dataStore, now: func() time.Time { return now }}

			attorney, err := attorneyStore.Create(ctx, "session-id", "attorney-id", isReplacement)
			assert.Nil(t, err)
			assert.Equal(t, details, attorney)
		})
	}
}

func TestAttorneyStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &attorneyStore{dataStore: nil, now: nil}

	_, err := attorneyStore.Create(ctx, "session-id", "attorney-id", false)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestAttorneyStoreCreateWhenSessionDataMissing(t *testing.T) {
	testcases := map[string]*page.SessionData{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), sessionData)

			attorneyStore := &attorneyStore{}

			_, err := attorneyStore.Create(ctx, "session-id", "attorney-id", false)
			assert.NotNil(t, err)
		})
	}
}

func TestAttorneyStoreCreateWhenCreateError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
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

			attorneyStore := &attorneyStore{dataStore: dataStore, now: func() time.Time { return now }}

			_, err := attorneyStore.Create(ctx, "session-id", "attorney-id", false)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestAttorneyStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	attorney := &actor.AttorneyProvidedDetails{LpaID: "123"}

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGetAllByGsi(ctx, "ActorIndex", "#ATTORNEY#session-id",
			[]any{attorney}, nil)

	attorneyStore := &attorneyStore{dataStore: dataStore, now: nil}

	attorneys, err := attorneyStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*actor.AttorneyProvidedDetails{attorney}, attorneys)
}

func TestAttorneyStoreGetAllWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &attorneyStore{}

	_, err := attorneyStore.GetAll(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestAttorneyStoreGetAllWhenMissingSessionID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	attorneyStore := &attorneyStore{}

	_, err := attorneyStore.GetAll(ctx)
	assert.NotNil(t, err)
}

func TestAttorneyStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#ATTORNEY#456",
			&actor.AttorneyProvidedDetails{LpaID: "123"}, nil)

	attorneyStore := &attorneyStore{dataStore: dataStore, now: nil}

	attorney, err := attorneyStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.AttorneyProvidedDetails{LpaID: "123"}, attorney)
}

func TestAttorneyStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &attorneyStore{dataStore: nil, now: nil}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestAttorneyStoreGetMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "456"})

	attorneyStore := &attorneyStore{}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, errors.New("attorneyStore.Get requires LpaID and SessionID"), err)
}

func TestAttorneyStoreGetMissingSessionIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	attorneyStore := &attorneyStore{}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, errors.New("attorneyStore.Get requires LpaID and SessionID"), err)
}

func TestAttorneyStoreGetOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#ATTORNEY#456",
			&actor.AttorneyProvidedDetails{LpaID: "123"}, expectedError)

	attorneyStore := &attorneyStore{dataStore: dataStore, now: nil}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestAttorneyStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, &actor.AttorneyProvidedDetails{PK: "LPA#123", SK: "#ATTORNEY#456", LpaID: "123", UpdatedAt: now}).
		Return(nil)

	attorneyStore := &attorneyStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &actor.AttorneyProvidedDetails{PK: "LPA#123", SK: "#ATTORNEY#456", LpaID: "123"})
	assert.Nil(t, err)
}

func TestAttorneyStorePutOnError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, &actor.AttorneyProvidedDetails{PK: "LPA#123", SK: "#ATTORNEY#456", LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	attorneyStore := &attorneyStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &actor.AttorneyProvidedDetails{PK: "LPA#123", SK: "#ATTORNEY#456", LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
