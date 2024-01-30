package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

	assert.Equal(t, &shareCodeStore{dynamoClient: client}, NewShareCodeStore(client))
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
