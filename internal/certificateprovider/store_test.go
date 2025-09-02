package certificateprovider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoreCreate(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "lpa-id", SessionID: "session-id"})
	uid := actoruid.New()
	details := &certificateproviderdata.Provided{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.CertificateProviderKey("session-id"), LpaID: "lpa-id", UpdatedAt: testNow, UID: uid, Email: "a@b.com"}

	link := accesscodedata.Link{
		PK:          dynamo.AccessKey(dynamo.CertificateProviderAccessKey("share-key")),
		SK:          dynamo.AccessSortKey(dynamo.MetadataKey("share-key")),
		ActorUID:    uid,
		UpdatedAt:   testNow,
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	expectedTransaction := &dynamo.Transaction{
		Creates: []any{
			dynamo.Keys{PK: details.PK, SK: dynamo.ReservedKey(dynamo.CertificateProviderKey)},
			details,
			dashboarddata.LpaLink{
				PK:        dynamo.LpaKey("lpa-id"),
				SK:        dynamo.SubKey("session-id"),
				DonorKey:  link.LpaOwnerKey,
				UID:       uid,
				ActorType: actor.TypeCertificateProvider,
				UpdatedAt: testNow,
			},
		},
		Deletes: []dynamo.Keys{{PK: link.PK, SK: link.SK}},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, expectedTransaction).
		Return(nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	certificateProvider, err := certificateProviderStore.Create(ctx, link, "a@b.com")
	assert.Nil(t, err)
	assert.Equal(t, details, certificateProvider)
}

func TestStoreCreateWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Create(ctx, accesscodedata.Link{}, "")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreCreateWhenSessionMissingRequiredData(t *testing.T) {
	testcases := map[string]*appcontext.Session{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(ctx, sessionData)

			certificateProviderStore := &Store{}

			_, err := certificateProviderStore.Create(ctx, accesscodedata.Link{}, "")
			assert.NotNil(t, err)
		})
	}
}

func TestStoreCreateWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := certificateProviderStore.Create(ctx, accesscodedata.Link{
		PK: dynamo.AccessKey(dynamo.CertificateProviderAccessKey("123")),
		SK: dynamo.AccessSortKey(dynamo.MetadataKey("123")),
	}, "")
	assert.Equal(t, expectedError, err)
}

func TestStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &certificateproviderdata.Provided{LpaID: "123"}, nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &certificateproviderdata.Provided{LpaID: "123"}, certificateProvider)
}

func TestStoreGetAnyWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreGetAnyMissingLpaIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.GetAny requires LpaID"), err)
}

func TestStoreGetAnyOnError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &certificateproviderdata.Provided{LpaID: "123"}, expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &certificateproviderdata.Provided{LpaID: "123"}, nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &certificateproviderdata.Provided{LpaID: "123"}, certificateProvider)
}

func TestStoreGetWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreGetMissingLpaIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "456"})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestStoreGetMissingSessionIDInSession(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestStoreGetOnError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &certificateproviderdata.Provided{LpaID: "123"}, expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestStoreOneByUID(t *testing.T) {
	provided := &certificateproviderdata.Provided{LpaID: "lpa-id"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("lpa-id")}, nil)
	dynamoClient.EXPECT().
		OneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.CertificateProviderKey(""), mock.Anything).
		Return(nil).
		SetData(provided)

	store := &Store{dynamoClient: dynamoClient}

	result, err := store.OneByUID(ctx, "M-1111-2222-3333")
	assert.Nil(t, err)
	assert.Equal(t, provided, result)
}

func TestStoreOneByUIDWhenNotFound(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, nil)

	store := &Store{dynamoClient: dynamoClient}

	_, err := store.OneByUID(ctx, "M-1111-2222-3333")
	assert.ErrorIs(t, err, dynamo.NotFoundError{})
}

func TestStoreOneByUIDWhenError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, expectedError)

	store := &Store{dynamoClient: dynamoClient}

	_, err := store.OneByUID(ctx, "M-1111-2222-3333")
	assert.ErrorIs(t, err, expectedError)
}

func TestStorePut(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: testNow}).
		Return(nil)

	certificateProviderStore := &Store{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := certificateProviderStore.Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Nil(t, err)
}

func TestStorePutOnError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: testNow}).
		Return(expectedError)

	certificateProviderStore := &Store{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := certificateProviderStore.Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestStoreDelete(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456")},
				{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("456")},
				{PK: dynamo.LpaKey("123"), SK: dynamo.ReservedKey(dynamo.CertificateProviderKey)},
			},
		}).
		Return(nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient}

	err := certificateProviderStore.Delete(ctx)
	assert.Nil(t, err)
}

func TestStoreDeleteWhenMissingSessionValues(t *testing.T) {
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

			certificateProviderStore := &Store{}

			err := certificateProviderStore.Delete(ctx)
			assert.Error(t, err)
		})
	}
}

func TestStoreDeleteWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient}

	err := certificateProviderStore.Delete(ctx)
	assert.Equal(t, fmt.Errorf("error deleting certificate provider: %w", expectedError), err)
}
