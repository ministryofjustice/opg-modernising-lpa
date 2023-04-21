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
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProvider{LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: func() time.Time { return now }}

	certificateProvider, err := certificateProviderStore.Create(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProvider{LpaID: "123", UpdatedAt: now}, certificateProvider)
}

func TestCertificateProviderStoreCreateWhenPutError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})
	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProvider{LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: func() time.Time { return now }}

	_, err := certificateProviderStore.Create(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProvider{LpaID: "123"}, nil)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	certificateProvider, err := certificateProviderStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.CertificateProvider{LpaID: "123"}, certificateProvider)
}

func TestCertificateProviderStoreGetMissingLpaIDInSessionData(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	certificateProviderStore := &certificateProviderStore{}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, errors.New("certificateProviderStore.Get requires LpaID to retrieve"), err)
}

func TestCertificateProviderStoreGetOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"})

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "LPA#123", "#CERTIFICATE_PROVIDER#", &actor.CertificateProvider{LpaID: "123"}, expectedError)

	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: nil}

	_, err := certificateProviderStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestCertificateProviderStorePut(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProvider{LpaID: "123", UpdatedAt: now}).
		Return(nil)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProvider{LpaID: "123"})

	assert.Nil(t, err)
}

func TestCertificateProviderStorePutOnError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	now := time.Now()

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, "LPA#123", "#CERTIFICATE_PROVIDER#456", &actor.CertificateProvider{LpaID: "123", UpdatedAt: now}).
		Return(expectedError)

	certificateProviderStore := &certificateProviderStore{
		dataStore: dataStore,
		now:       func() time.Time { return now },
	}

	err := certificateProviderStore.Put(ctx, &actor.CertificateProvider{LpaID: "123"})

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

		err := certificateProviderStore.Put(ctx, &actor.CertificateProvider{})
		assert.Equal(t, errors.New("certificateProviderStore.Put requires LpaID and SessionID to retrieve"), err)
	}
}
