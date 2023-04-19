package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/mock"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	lpa := &page.Lpa{ID: "1", CertificateProviderID: "10100000"}

	lpaPk, _ := attributevalue.Marshal("DONOR#sesh-id")
	lpaSk, _ := attributevalue.Marshal("#LPA#1")
	lpaData, _ := attributevalue.Marshal(lpa)

	update := &types.Update{
		Key:              map[string]types.AttributeValue{"PK": lpaPk, "SK": lpaSk},
		UpdateExpression: aws.String("SET #Data = :lpa"),
		ExpressionAttributeNames: map[string]string{
			"#Data": "Data",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lpa": lpaData,
		},
	}

	dataStore := newMockDataStore(t)
	dataStore.
		On("PutTransact", ctx, "CERTIFICATE_PROVIDER#10100000", "#LPA#1", mock.Anything, update).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	certificateProvider, err := certificateProviderStore.Create(ctx, lpa, "sesh-id")
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProvider{ID: "10100000", LpaID: "1"}, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenPutTransactError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	lpa := &page.Lpa{ID: "1", CertificateProviderID: "10100000"}

	lpaPk, _ := attributevalue.Marshal("DONOR#sesh-id")
	lpaSk, _ := attributevalue.Marshal("#LPA#1")
	lpaData, _ := attributevalue.Marshal(lpa)

	update := &types.Update{
		Key:              map[string]types.AttributeValue{"PK": lpaPk, "SK": lpaSk},
		UpdateExpression: aws.String("SET #Data = :lpa"),
		ExpressionAttributeNames: map[string]string{
			"#Data": "Data",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lpa": lpaData,
		},
	}

	dataStore := newMockDataStore(t)
	dataStore.
		On("PutTransact", ctx, "CERTIFICATE_PROVIDER#10100000", "#LPA#1", mock.Anything, update).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	certificateProvider, err := certificateProviderStore.Create(ctx, lpa, "sesh-id")
	assert.Equal(t, expectedError, err)
	assert.Equal(t, &actor.CertificateProvider{ID: "10100000", LpaID: "1"}, certificateProvider)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", CertificateProviderID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "CERTIFICATE_PROVIDER#456", "#LPA#123", &actor.CertificateProvider{ID: "456", LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProvider{ID: "456", LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetWhenMissingLpaID(t *testing.T) {
	testCases := map[string]struct {
		sessionData *page.SessionData
	}{
		"missing LpaID":                 {sessionData: &page.SessionData{CertificateProviderID: "456"}},
		"missing CertificateProviderID": {sessionData: &page.SessionData{LpaID: "123"}},
		"missing both":                  {sessionData: &page.SessionData{}},
	}

	for _, tc := range testCases {
		ctx := page.ContextWithSessionData(context.Background(), tc.sessionData)

		certificateProviderStore := &certificateProviderStore{dataStore: nil, randomInt: func(x int) int { return x }}

		_, err := certificateProviderStore.Get(ctx)
		assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID and CertificateProviderId to retrieve"), err)
	}
}

func TestCertificateProviderStoreGetOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", CertificateProviderID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "CERTIFICATE_PROVIDER#456", "#LPA#123", &actor.CertificateProvider{ID: "456", LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, randomInt: func(x int) int { return x }}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "CERTIFICATE_PROVIDER#123", "#LPA#456", &actor.CertificateProvider{ID: "123", LpaID: "456", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		randomInt: func(x int) int { return x },
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProvider{ID: "123", LpaID: "456"})

	assert.Nil(t, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "CERTIFICATE_PROVIDER#123", "#LPA#456", &actor.CertificateProvider{ID: "123", LpaID: "456", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		randomInt: func(x int) int { return x },
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProvider{ID: "123", LpaID: "456"})

	assert.Equal(t, expectedError, err)
}
