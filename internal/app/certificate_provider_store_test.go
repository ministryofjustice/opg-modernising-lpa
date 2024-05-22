package app

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	uid := actoruid.New()
	now := time.Now()
	nowFormatted := now.Format(time.RFC3339Nano)
	details := &actor.CertificateProviderProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: now, UID: uid, Email: "a@b.com"}

	shareCode := actor.ShareCodeData{
		PK:          dynamo.ShareKey(dynamo.CertificateProviderShareKey("123")),
		SK:          dynamo.ShareSortKey(dynamo.MetadataKey("123")),
		ActorUID:    uid,
		UpdatedAt:   now,
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
	}

	marshalledCertificateProvider, _ := attributevalue.MarshalMap(details)

	expectedTransaction := &dynamo.Transaction{
		Puts: []*types.Put{
			{Item: marshalledCertificateProvider},
			{Item: map[string]types.AttributeValue{
				"PK":        &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK":        &types.AttributeValueMemberS{Value: "SUB#456"},
				"DonorKey":  &types.AttributeValueMemberS{Value: "DONOR#donor"},
				"ActorType": &types.AttributeValueMemberN{Value: "4"},
				"UpdatedAt": &types.AttributeValueMemberS{Value: nowFormatted},
			}},
		},
		Deletes: []*types.Delete{
			{Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: shareCode.PK.PK()},
				"SK": &types.AttributeValueMemberS{Value: shareCode.SK.SK()},
			}},
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, expectedTransaction).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	certificateProvider, err := certificateProviderStore.Create(ctx, shareCode, "a@b.com")
	assert.Nil(t, err)
	assert.Equal(t, details, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Create(ctx, actor.ShareCodeData{}, "")
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreCreateWhenSessionDataMissing(t *testing.T) {
	testcases := map[string]*page.SessionData{
		"LpaID":     {SessionID: "456"},
		"SessionID": {LpaID: "123"},
	}

	for name, sessionData := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), sessionData)

			certificateProviderStore := &certificateProviderStore{}

			_, err := certificateProviderStore.Create(ctx, actor.ShareCodeData{}, "")
			assert.NotNil(t, err)
		})
	}
}

func TestCertificateProviderStoreCreateWhenWriteTransactionError(t *testing.T) {
	now := time.Now()
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	_, err := certificateProviderStore.Create(ctx, actor.ShareCodeData{
		PK: dynamo.ShareKey(dynamo.CertificateProviderShareKey("123")),
		SK: dynamo.ShareSortKey(dynamo.MetadataKey("123")),
	}, "")
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetAnyWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetAnyMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.GetAny requires LpaID"), err)
}

func TestCertificateProviderStoreGetAnyOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "456"})

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetMissingSessionIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456"), &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.CertificateProviderProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Nil(t, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.CertificateProviderProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456"), LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreDelete(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey("456")).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient}

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
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: tc.lpaID, SessionID: tc.sessionID})

			certificateProviderStore := &certificateProviderStore{}

			err := certificateProviderStore.Delete(ctx)
			assert.Error(t, err)
		})
	}
}

func TestCertificateProviderStoreDeleteWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient}

	err := certificateProviderStore.Delete(ctx)
	assert.Equal(t, fmt.Errorf("error deleting certificate provider: %w", expectedError), err)
}
