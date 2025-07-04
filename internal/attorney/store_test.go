package attorney

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAttorneyStoreCreate(t *testing.T) {
	testcases := map[actor.Type]struct {
		replacement      bool
		trustCorporation bool
	}{
		actor.TypeAttorney:                    {},
		actor.TypeReplacementAttorney:         {replacement: true},
		actor.TypeTrustCorporation:            {trustCorporation: true},
		actor.TypeReplacementTrustCorporation: {replacement: true, trustCorporation: true},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})
			now := time.Now()
			uid := actoruid.New()
			details := &attorneydata.Provided{
				PK:                 dynamo.LpaKey("123"),
				SK:                 dynamo.AttorneyKey("456"),
				UID:                uid,
				LpaID:              "123",
				UpdatedAt:          now,
				IsReplacement:      tc.replacement,
				IsTrustCorporation: tc.trustCorporation,
				Email:              "a@example.com",
			}

			link := accesscodedata.Link{
				PK:                    dynamo.AccessKey(dynamo.AttorneyAccessKey("123")),
				SK:                    dynamo.ShareSortKey(dynamo.MetadataKey("123")),
				ActorUID:              uid,
				IsReplacementAttorney: tc.replacement,
				IsTrustCorporation:    tc.trustCorporation,
				UpdatedAt:             now,
				LpaOwnerKey:           dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
			}

			expectedTransaction := &dynamo.Transaction{
				Creates: []any{
					details,
					dashboarddata.LpaLink{
						PK:        dynamo.LpaKey("123"),
						SK:        dynamo.SubKey("456"),
						DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
						UID:       uid,
						ActorType: actorType,
						UpdatedAt: now,
					},
				},
				Deletes: []dynamo.Keys{
					{
						PK: dynamo.AccessKey(dynamo.AttorneyAccessKey("123")),
						SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
					},
				},
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, expectedTransaction).
				Return(nil)

			attorneyStore := Store{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			attorney, err := attorneyStore.Create(ctx, link, "a@example.com")
			assert.Nil(t, err)
			assert.Equal(t, details, attorney)
		})
	}
}

func TestAttorneyStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &Store{dynamoClient: nil, now: nil}

	_, err := attorneyStore.Create(ctx, accesscodedata.Link{}, "")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestAttorneyStoreCreateWhenSessionMissingRequiredData(t *testing.T) {
	testcases := map[string]*appcontext.Session{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), sessionData)

			attorneyStore := &Store{}

			_, err := attorneyStore.Create(ctx, accesscodedata.Link{}, "")
			assert.NotNil(t, err)
		})
	}
}

func TestAttorneyStoreCreateWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	attorneyStore := &Store{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	_, err := attorneyStore.Create(ctx, accesscodedata.Link{
		PK: dynamo.AccessKey(dynamo.AttorneyAccessKey("123")),
		SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
	}, "")
	assert.Equal(t, expectedError, err)
}

func TestAttorneyStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.AttorneyKey("456"),
			&attorneydata.Provided{LpaID: "123"}, nil)

	attorneyStore := &Store{dynamoClient: dynamoClient, now: nil}

	attorney, err := attorneyStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &attorneydata.Provided{LpaID: "123"}, attorney)
}

func TestAttorneyStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	attorneyStore := &Store{dynamoClient: nil, now: nil}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestAttorneyStoreGetMissingLpaIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "456"})

	attorneyStore := &Store{}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, errors.New("attorneyStore.Get requires LpaID and SessionID"), err)
}

func TestAttorneyStoreGetMissingSessionIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	attorneyStore := &Store{}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, errors.New("attorneyStore.Get requires LpaID and SessionID"), err)
}

func TestAttorneyStoreGetOnError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.AttorneyKey("456"),
			&attorneydata.Provided{LpaID: "123"}, expectedError)

	attorneyStore := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := attorneyStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestAttorneyStoreAll(t *testing.T) {
	ctx := context.Background()
	pk := dynamo.LpaKey("an-lpa")
	expected := []*attorneydata.Provided{{LpaID: "lpa-id"}}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByPartialSK(ctx, pk, dynamo.AttorneyKey(""), mock.Anything).
		Return(expectedError).
		SetData(expected)

	attorneyStore := &Store{dynamoClient: dynamoClient, now: nil}

	attorney, err := attorneyStore.All(ctx, pk)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, expected, attorney)
}

func TestAttorneyStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &attorneydata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(nil)

	attorneyStore := &Store{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &attorneydata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123"})
	assert.Nil(t, err)
}

func TestAttorneyStorePutOnError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &attorneydata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	attorneyStore := &Store{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := attorneyStore.Put(ctx, &attorneydata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.AttorneyKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestAttorneyStoreDelete(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.LpaKey("123"), dynamo.AttorneyKey("456")).
		Return(nil)

	attorneyStore := &Store{dynamoClient: dynamoClient}

	err := attorneyStore.Delete(ctx)
	assert.Nil(t, err)
}

func TestAttorneyStoreDeleteWhenSessionErrors(t *testing.T) {
	attorneyStore := &Store{}

	err := attorneyStore.Delete(ctx)
	assert.Error(t, err)
}

func TestAttorneyStoreDeleteWhenMissingSessionValues(t *testing.T) {
	testcases := map[string]struct {
		lpaID     string
		sessionID string
	}{
		"missing LpaID": {
			sessionID: "456",
		},
		"missing SessionID": {
			lpaID: "123",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: tc.lpaID, SessionID: tc.sessionID})

			attorneyStore := &Store{}

			err := attorneyStore.Delete(ctx)
			assert.Error(t, err)
		})
	}
}

func TestAttorneyStoreDeleteWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	attorneyStore := &Store{dynamoClient: dynamoClient}

	err := attorneyStore.Delete(ctx)
	assert.Equal(t, fmt.Errorf("error deleting attorney: %w", expectedError), err)
}
