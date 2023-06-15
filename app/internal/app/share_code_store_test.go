package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
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

			dataStore := newMockDataStore(t)
			dataStore.
				ExpectGet(ctx, tc.pk, "#METADATA#123",
					data, nil)

			shareCodeStore := &shareCodeStore{dataStore: dataStore}

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

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(ctx, "ATTORNEYSHARE#123", "#METADATA#123",
			data, expectedError)

	shareCodeStore := &shareCodeStore{dataStore: dataStore}

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
			data := actor.ShareCodeData{LpaID: "lpa-id"}

			dataStore := newMockDataStore(t)
			dataStore.
				On("Put", ctx, tc.pk, "#METADATA#123", data).
				Return(nil)

			shareCodeStore := &shareCodeStore{dataStore: dataStore}

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
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123", SessionID: "456"})

	dataStore := newMockDataStore(t)
	dataStore.
		On("Put", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeStore := &shareCodeStore{dataStore: dataStore}

	err := shareCodeStore.Put(ctx, actor.TypeAttorney, "123", actor.ShareCodeData{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
