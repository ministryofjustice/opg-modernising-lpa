package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestAttorneyStoreCreate(t *testing.T) {
	testcases := map[string]struct {
		replacement      bool
		trustCorporation bool
	}{
		"attorney":                      {},
		"replacement":                   {replacement: true},
		"trust corporation":             {trustCorporation: true},
		"replacement trust corporation": {replacement: true, trustCorporation: true},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
			now := time.Now()
			uid := actoruid.New()
			details := &actor.AttorneyProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), UID: uid, LpaID: "123", UpdatedAt: now, IsReplacement: tc.replacement, IsTrustCorporation: tc.trustCorporation, Email: "a@example.com"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, details).
				Return(nil)
			dynamoClient.EXPECT().
				Create(ctx, lpaLink{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("456"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), ActorType: actor.TypeAttorney, UpdatedAt: now}).
				Return(nil)

			attorneyStore := &attorneyStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			attorney, err := attorneyStore.Create(ctx, dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), uid, tc.replacement, tc.trustCorporation, "a@example.com")
			assert.Nil(t, err)
			assert.Equal(t, details, attorney)
		})
	}
}

func TestAttorneyStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &attorneyStore{dynamoClient: nil, now: nil}

	_, err := attorneyStore.Create(ctx, dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), actoruid.New(), false, false, "")
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

			_, err := attorneyStore.Create(ctx, dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), actoruid.New(), false, false, "")
			assert.NotNil(t, err)
		})
	}
}

func TestAttorneyStoreCreateWhenCreateError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	now := time.Now()

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"certificate provider record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"link record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDataStore(t)

			attorneyStore := &attorneyStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			_, err := attorneyStore.Create(ctx, dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), actoruid.New(), false, false, "")
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestAttorneyStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.AttorneyKey("456"),
			&actor.AttorneyProvidedDetails{LpaID: "123"}, nil)

	attorneyStore := &attorneyStore{dynamoClient: dynamoClient, now: nil}

	attorney, err := attorneyStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.AttorneyProvidedDetails{LpaID: "123"}, attorney)
}

func TestAttorneyStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &attorneyStore{dynamoClient: nil, now: nil}

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

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.AttorneyKey("456"),
			&actor.AttorneyProvidedDetails{LpaID: "123"}, expectedError)

	attorneyStore := &attorneyStore{dynamoClient: dynamoClient, now: nil}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestAttorneyStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.AttorneyProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(nil)

	attorneyStore := &attorneyStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &actor.AttorneyProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123"})
	assert.Nil(t, err)
}

func TestAttorneyStorePutOnError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.AttorneyProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	attorneyStore := &attorneyStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &actor.AttorneyProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
