package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShareCodeStoreGet(t *testing.T) {
	testcases := map[string]struct {
		t  actor.Type
		pk string
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: "ATTORNEYSHARE#123",
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: "ATTORNEYSHARE#123",
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: "CERTIFICATEPROVIDERSHARE#123",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := actor.ShareCodeData{LpaID: "lpa-id"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOne(ctx, tc.pk, "#METADATA#123",
					data, nil)

			shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

			result, err := shareCodeStore.Get(ctx, tc.t, "123")
			assert.Nil(t, err)
			assert.Equal(t, data, result)
		})
	}
}

func TestShareCodeStoreGetForBadActorType(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &shareCodeStore{}

	_, err := shareCodeStore.Get(ctx, actor.TypeDonor, "123")
	assert.NotNil(t, err)
}

func TestShareCodeStoreGetOnError(t *testing.T) {
	ctx := context.Background()
	data := actor.ShareCodeData{LpaID: "lpa-id"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "ATTORNEYSHARE#123", "#METADATA#123",
			data, expectedError)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	_, err := shareCodeStore.Get(ctx, actor.TypeAttorney, "123")
	assert.Equal(t, expectedError, err)
}

func TestShareCodeStorePut(t *testing.T) {
	testcases := map[string]struct {
		actor actor.Type
		pk    string
	}{
		"attorney": {
			actor: actor.TypeAttorney,
			pk:    "ATTORNEYSHARE#123",
		},
		"replacement attorney": {
			actor: actor.TypeReplacementAttorney,
			pk:    "ATTORNEYSHARE#123",
		},
		"certificate provider": {
			actor: actor.TypeCertificateProvider,
			pk:    "CERTIFICATEPROVIDERSHARE#123",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := actor.ShareCodeData{PK: tc.pk, SK: "#METADATA#123", LpaID: "lpa-id"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Put(ctx, data).
				Return(nil)

			shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

			err := shareCodeStore.Put(ctx, tc.actor, "123", data)
			assert.Nil(t, err)
		})
	}
}

func TestShareCodeStorePutForBadActorType(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &shareCodeStore{}

	err := shareCodeStore.Put(ctx, actor.TypePersonToNotify, "123", actor.ShareCodeData{})
	assert.NotNil(t, err)
}

func TestShareCodeStorePutOnError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	err := shareCodeStore.Put(ctx, actor.TypeAttorney, "123", actor.ShareCodeData{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestNewShareCodeStore(t *testing.T) {
	client := newMockDynamoClient(t)
	store := NewShareCodeStore(client)

	assert.Equal(t, client, store.dynamoClient)
	assert.NotNil(t, store.now)
}

func TestShareCodeStoreGetDonor(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{
		OrganisationID: "org-id",
		LpaID:          "lpa-id",
	})
	data := actor.ShareCodeData{LpaID: "lpa-id"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, "DONORINVITE#org-id#lpa-id",
			data, expectedError)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	result, err := shareCodeStore.GetDonor(ctx)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, data, result)
}

func TestShareCodeStoreGetDonorWithSessionMissing(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &shareCodeStore{}

	_, err := shareCodeStore.GetDonor(ctx)
	assert.NotNil(t, err)
}

func TestShareCodeStorePutDonor(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, actor.ShareCodeData{
			PK:        "DONORSHARE#123",
			SK:        "DONORINVITE#org-id#lpa-id",
			SessionID: "org-id",
			LpaID:     "lpa-id",
			UpdatedAt: testNow,
		}).
		Return(nil)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient, now: testNowFn}

	err := shareCodeStore.PutDonor(ctx, "123", actor.ShareCodeData{SessionID: "org-id", LpaID: "lpa-id"})
	assert.Nil(t, err)
}

func TestShareCodeStoreDelete(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, "a-pk", "a-sk").
		Return(nil)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	err := shareCodeStore.Delete(ctx, actor.ShareCodeData{LpaID: "123", PK: "a-pk", SK: "a-sk"})
	assert.Nil(t, err)
}

func TestShareCodeStoreDeleteOnError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	err := shareCodeStore.Delete(ctx, actor.ShareCodeData{})
	assert.Equal(t, expectedError, err)
}
