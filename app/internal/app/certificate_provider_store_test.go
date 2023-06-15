package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Create", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}).
		Return(nil)
	dataStore.
		On("Create", ctx, "LPA#123", "#SUB#456", "#DONOR#session-id|CERTIFICATE_PROVIDER").
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: func() time.Time { return now }}

	certificateProvider, err := certificateProviderStore.Create(ctx, "session-id")
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dataStore: nil, now: nil}

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

	testcases := map[string]func(*testing.T) *mockDataStore{
		"certificate provider record": func(t *testing.T) *mockDataStore {
			dataStore := newMockDataStore(t)
			dataStore.
				On("Create", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}).
				Return(expectedError)

			return dataStore
		},
		"link record": func(t *testing.T) *mockDataStore {
			dataStore := newMockDataStore(t)
			dataStore.
				On("Create", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}).
				Return(nil)
			dataStore.
				On("Create", ctx, "LPA#123", "#SUB#456", "#DONOR#session-id|CERTIFICATE_PROVIDER").
				Return(expectedError)

			return dataStore
		},
	}

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dataStore := makeMockDataStore(t)

			certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: func() time.Time { return now }}

			_, err := certificateProviderStore.Create(ctx, "session-id")
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestCertificateProviderStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	certificateProvider := &actor.CertificateProviderProvidedDetails{LpaID: "123"}

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGetAllByGsi(ctx, "ActorIndex", "#CERTIFICATE_PROVIDER#session-id",
			[]map[string]any{{"Data": certificateProvider}}, nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

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

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGetOneByPartialSk(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetAnyWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dataStore: nil, now: nil}

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

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGetOneByPartialSk(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	_, err := certificateProviderStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProviderProvidedDetails{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dataStore: nil, now: nil}

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

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{LpaID: "123"})

	assert.Nil(t, err)
}

func TestCertificateProviderStorePutWhenSessionMissing(t *testing.T) {
	ctx := context.Background()

	certificateProviderStore := &certificateProviderStore{dataStore: nil, now: nil}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{})
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProviderProvidedDetails{LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{LpaID: "123"})

	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePutMissingRequiredSessionData(t *testing.T) {
	testCases := map[string]struct {
		sessionData *page.SessionData
	}{
		"missing LpaID":     {sessionData: &page.SessionData{SessionID: "456"}},
		"missing SessionID": {sessionData: &page.SessionData{LpaID: "123"}},
		"missing both":      {sessionData: &page.SessionData{}},
	}

	for _, tc := range testCases {
		ctx := page.ContextWithSessionData(context.Background(), tc.sessionData)

		certificateProviderStore := &certificateProviderStore{dataStore: nil}

		err := certificateProviderStore.Put(ctx, &actor.CertificateProviderProvidedDetails{})
		assert.Equal(t, errors.New("certificateProviderStore.Put requires LpaID and SessionID"), err)
	}
}
