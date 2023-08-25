package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	now := time.Now()
	details := &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, details).
		Return(nil)
	dynamoClient.
		On("Create", ctx, lpaLink{PK: "LPA#123", SK: "#SUB#456", DonorKey: "#DONOR#session-id", ActorType: actor.TypeCertificateProvider}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	certificateProvider, err := certificateProviderStore.Create(ctx, "session-id")
	assert.Nil(t, err)
	assert.Equal(t, details, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Create(ctx, "session-id")
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

			_, err := certificateProviderStore.Create(ctx, "session-id")
			assert.NotNil(t, err)
		})
	}
}

func TestCertificateProviderStoreCreateWhenCreateError(t *testing.T) {
	now := time.Now()
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"certificate provider record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				On("Create", ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"link record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				On("Create", ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.
				On("Create", ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDataStore(t)

			certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			_, err := certificateProviderStore.Create(ctx, "session-id")
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestCertificateProviderStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	certificateProvider := &actor.CertificateProviderProvidedDetails{LpaID: "123"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectGetAllByGsi(ctx, "ActorIndex", "#CERTIFICATE_PROVIDER#session-id",
			[]any{certificateProvider}, nil)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	certificateProviders, err := certificateProviderStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*actor.CertificateProviderProvidedDetails{certificateProvider}, certificateProviders)
}

func TestCertificateProviderStoreGetAllWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.GetAll(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetAllWhenMissingSessionID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.GetAll(ctx)
	assert.NotNil(t, err)
}

func TestCertificateProviderStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectGetOneByPartialSk(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

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
		ExpectGetOneByPartialSk(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

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
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123"})
	assert.Nil(t, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
