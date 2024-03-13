package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	uid := actoruid.New()
	now := time.Now()
	details := &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now, UID: uid}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, details).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, lpaLink{PK: "LPA#123", SK: "#SUB#456", DonorKey: "#DONOR#session-id", ActorType: actor.TypeCertificateProvider, UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	certificateProvider, err := certificateProviderStore.Create(ctx, "session-id", uid)
	assert.Nil(t, err)
	assert.Equal(t, details, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &CertificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Create(ctx, "session-id", actoruid.New())
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

			certificateProviderStore := &CertificateProviderStore{}

			_, err := certificateProviderStore.Create(ctx, "session-id", actoruid.New())
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

			certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			_, err := certificateProviderStore.Create(ctx, "session-id", actoruid.New())
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestCertificateProviderStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetAnyWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &CertificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetAnyMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	certificateProviderStore := &CertificateProviderStore{}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, errors.New("CertificateProviderStore.GetAny requires LpaID"), err)
}

func TestCertificateProviderStoreGetAnyOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &CertificateProviderStore{dynamoClient: nil, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStoreGetMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "456"})

	certificateProviderStore := &CertificateProviderStore{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("CertificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetMissingSessionIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	certificateProviderStore := &CertificateProviderStore{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("CertificateProviderStore.Get requires LpaID and SessionID"), err)
}

func TestCertificateProviderStoreGetOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &CertificateProviderStore{
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
	dynamoClient.EXPECT().
		Put(ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &CertificateProviderStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{PK: "LPA#123", SK: "#CERTIFICATE_PROVIDER#456", LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreCreatePaper(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()
	now := time.Now()
	details := &actor.CertificateProviderProvidedDetails{
		PK:        "LPA#123",
		SK:        "#CERTIFICATE_PROVIDER#" + uid.String(),
		LpaID:     "123",
		UpdatedAt: now,
		UID:       uid,
		Certificate: actor.Certificate{
			AgreeToStatement: true,
			Agreed:           now,
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, details).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, lpaLink{
			PK:        "LPA#123",
			SK:        "#SUB#" + uid.String(),
			DonorKey:  "#DONOR#session-id",
			ActorType: actor.TypeCertificateProvider,
			UpdatedAt: now,
		}).
		Return(nil)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := certificateProviderStore.CreatePaper(ctx, "123", uid, "session-id")
	assert.Nil(t, err)
}

func TestCertificateProviderStoreCreatePaperWhenDynamoCreateCertificateProviderError(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := certificateProviderStore.CreatePaper(ctx, "123", uid, "session-id")
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreCreatePaperWhenDynamoCreateLpaLinkError(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(nil).
		Once()

	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	certificateProviderStore := &CertificateProviderStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := certificateProviderStore.CreatePaper(ctx, "123", uid, "session-id")
	assert.Equal(t, expectedError, err)
}
