package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShareCodeStoreGet(t *testing.T) {
	testcases := map[string]struct {
		t  actor.Type
		pk dynamo.ShareKeyType
	}{
		"attorney": {
			t:  actor.TypeAttorney,
			pk: dynamo.ShareKey(dynamo.AttorneyShareKey("123")),
		},
		"replacement attorney": {
			t:  actor.TypeReplacementAttorney,
			pk: dynamo.ShareKey(dynamo.AttorneyShareKey("123")),
		},
		"certificate provider": {
			t:  actor.TypeCertificateProvider,
			pk: dynamo.ShareKey(dynamo.CertificateProviderShareKey("123")),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := actor.ShareCodeData{LpaKey: "lpa-id"}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPK(ctx, tc.pk,
					data, nil)

			shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

			result, err := shareCodeStore.Get(ctx, tc.t, "123")
			assert.Nil(t, err)
			assert.Equal(t, data, result)
		})
	}
}

func TestShareCodeStoreGetWhenLinked(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPK(ctx, dynamo.ShareKey(dynamo.DonorShareKey("123")),
			actor.ShareCodeData{LpaLinkedAt: time.Now()}, nil)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	result, err := shareCodeStore.Get(ctx, actor.TypeDonor, "123")
	assert.Equal(t, dynamo.NotFoundError{}, err)
	assert.Equal(t, actor.ShareCodeData{}, result)
}

func TestShareCodeStoreGetForBadActorType(t *testing.T) {
	ctx := context.Background()
	shareCodeStore := &shareCodeStore{}

	_, err := shareCodeStore.Get(ctx, actor.TypeIndependentWitness, "123")
	assert.NotNil(t, err)
}

func TestShareCodeStoreGetOnError(t *testing.T) {
	ctx := context.Background()
	data := actor.ShareCodeData{LpaKey: "lpa-id"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPK(ctx, dynamo.ShareKey(dynamo.AttorneyShareKey("123")),
			data, expectedError)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	_, err := shareCodeStore.Get(ctx, actor.TypeAttorney, "123")
	assert.Equal(t, expectedError, err)
}

func TestShareCodeStoreLinked(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, actor.ShareCodeData{
			LpaLinkedTo: "email",
			LpaLinkedAt: testNow,
		}).
		Return(expectedError)

	store := &shareCodeStore{dynamoClient: dynamoClient, now: testNowFn}
	err := store.Linked(ctx, actor.ShareCodeData{}, "email")

	assert.Equal(t, expectedError, err)
}

func TestShareCodeStorePut(t *testing.T) {
	testcases := map[string]struct {
		actor actor.Type
		pk    dynamo.ShareKeyType
	}{
		"attorney": {
			actor: actor.TypeAttorney,
			pk:    dynamo.ShareKey(dynamo.AttorneyShareKey("123")),
		},
		"replacement attorney": {
			actor: actor.TypeReplacementAttorney,
			pk:    dynamo.ShareKey(dynamo.AttorneyShareKey("123")),
		},
		"certificate provider": {
			actor: actor.TypeCertificateProvider,
			pk:    dynamo.ShareKey(dynamo.CertificateProviderShareKey("123")),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data := actor.ShareCodeData{PK: tc.pk, SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")), LpaKey: "lpa-id"}

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

	err := shareCodeStore.Put(ctx, actor.TypeAttorney, "123", actor.ShareCodeData{LpaKey: "123"})
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
	data := actor.ShareCodeData{LpaKey: "lpa-id"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.DonorInviteKey("org-id", "lpa-id"),
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
			PK:          dynamo.ShareKey(dynamo.DonorShareKey("123")),
			SK:          dynamo.ShareSortKey(dynamo.DonorInviteKey("org-id", "lpa-id")),
			LpaOwnerKey: "org-id",
			LpaKey:      "lpa-id",
			UpdatedAt:   testNow,
		}).
		Return(nil)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient, now: testNowFn}

	err := shareCodeStore.PutDonor(ctx, "123", actor.ShareCodeData{LpaOwnerKey: "org-id", LpaKey: "lpa-id"})
	assert.Nil(t, err)
}

func TestShareCodeStoreDelete(t *testing.T) {
	ctx := context.Background()
	pk := dynamo.ShareKey(dynamo.AttorneyShareKey("a-pk"))
	sk := dynamo.ShareSortKey(dynamo.MetadataKey("a-sk"))

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, pk, sk).
		Return(nil)

	shareCodeStore := &shareCodeStore{dynamoClient: dynamoClient}

	err := shareCodeStore.Delete(ctx, actor.ShareCodeData{LpaKey: "123", PK: pk, SK: sk})
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

func TestShareCodeKey(t *testing.T) {
	testcases := map[actor.Type]dynamo.PK{
		actor.TypeDonor:                       dynamo.ShareKey(dynamo.DonorShareKey("S")),
		actor.TypeAttorney:                    dynamo.ShareKey(dynamo.AttorneyShareKey("S")),
		actor.TypeReplacementAttorney:         dynamo.ShareKey(dynamo.AttorneyShareKey("S")),
		actor.TypeTrustCorporation:            dynamo.ShareKey(dynamo.AttorneyShareKey("S")),
		actor.TypeReplacementTrustCorporation: dynamo.ShareKey(dynamo.AttorneyShareKey("S")),
		actor.TypeCertificateProvider:         dynamo.ShareKey(dynamo.CertificateProviderShareKey("S")),
	}

	for actorType, prefix := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			pk, err := shareCodeKey(actorType, "S")
			assert.Nil(t, err)
			assert.Equal(t, prefix, pk)
		})
	}
}

func TestShareCodeKeyWhenUnknownType(t *testing.T) {
	_, err := shareCodeKey(actor.TypeAuthorisedSignatory, "S")
	assert.NotNil(t, err)
}
