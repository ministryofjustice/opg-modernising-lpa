package certificateproviderdata

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/temporary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "lpa-id", SessionID: "session-id"})
	uid := actoruid.New()
	details := &Provided{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.CertificateProviderKey("session-id"), LpaID: "lpa-id", UpdatedAt: testNow, UID: uid, Email: "a@b.com"}

	shareCode := sharecode.ShareCodeData{
		PK:          dynamo.ShareKey(dynamo.CertificateProviderShareKey("share-key")),
		SK:          dynamo.ShareSortKey(dynamo.MetadataKey("share-key")),
		ActorUID:    uid,
		UpdatedAt:   testNow,
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	expectedTransaction := &dynamo.Transaction{
		Creates: []any{
			details,
			temporary.LpaLink{
				PK:        dynamo.LpaKey("lpa-id"),
				SK:        dynamo.SubKey("session-id"),
				DonorKey:  shareCode.LpaOwnerKey,
				ActorType: temporary.ActorTypeCertificateProvider,
				UpdatedAt: testNow,
			},
		},
		Deletes: []dynamo.Keys{{PK: shareCode.PK, SK: shareCode.SK}},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, expectedTransaction).
		Return(nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	certificateProvider, err := certificateProviderStore.Create(ctx, shareCode, "a@b.com")
	assert.Nil(t, err)
	assert.Equal(t, details, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Create(ctx, sharecode.ShareCodeData{}, "")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestCertificateProviderStoreCreateWhenSessionDataMissing(t *testing.T) {
	testcases := map[string]*appcontext.SessionData{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSessionData(ctx, sessionData)

			certificateProviderStore := &Store{}

			_, err := certificateProviderStore.Create(ctx, sharecode.ShareCodeData{}, "")
			assert.NotNil(t, err)
		})
	}
}

func TestCertificateProviderStoreCreateWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	_, err := certificateProviderStore.Create(ctx, sharecode.ShareCodeData{
		PK: dynamo.ShareKey(dynamo.CertificateProviderShareKey("123")),
		SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
	}, "")
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &Provided{LpaID: "123"}, nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetAnyWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetAnyMissingLpaIDInSessionData(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.GetAny requires LpaID"), err)
}

func TestCertificateProviderStoreGetAnyOnError(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &Provided{LpaID: "123"}, expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &Provided{LpaID: "123"}, nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetWhenSessionMissing(t *testing.T) {
	certificateProviderStore := &Store{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetMissingLpaIDInSessionData(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{SessionID: "456"})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetMissingSessionIDInSessionData(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123"})

	certificateProviderStore := &Store{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetOnError(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &Provided{LpaID: "123"}, expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: testNow}).
		Return(nil)

	certificateProviderStore := &Store{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := certificateProviderStore.Put(ctx, &Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Nil(t, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: testNow}).
		Return(expectedError)

	certificateProviderStore := &Store{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := certificateProviderStore.Put(ctx, &Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreDelete(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456")).
		Return(nil)

	certificateProviderStore := &Store{dynamoClient: dynamoClient}

	err := certificateProviderStore.Delete(ctx)
	assert.Nil(t, err)
}

func TestCertificateProviderStoreDeleteWhenMissingSessionValues(t *testing.T) {
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
			ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: tc.lpaID, SessionID: tc.sessionID})

			certificateProviderStore := &Store{}

			err := certificateProviderStore.Delete(ctx)
			assert.Error(t, err)
		})
	}
}

func TestCertificateProviderStoreDeleteWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSessionData(context.Background(), &appcontext.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &Store{dynamoClient: dynamoClient}

	err := certificateProviderStore.Delete(ctx)
	assert.Equal(t, fmt.Errorf("error deleting certificate provider: %w", expectedError), err)
}
