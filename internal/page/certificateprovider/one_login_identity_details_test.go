package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOneLoginIdentityDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	certificateProvider := &actor.CertificateProviderProvidedDetails{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b"},
		LpaID:            "lpa-id",
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(certificateProvider, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &oneLoginIdentityDetailsData{
			App:                 testAppData,
			CertificateProvider: certificateProvider,
		}).
		Return(nil)

	err := OneLoginIdentityDetails(template.Execute, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetOneLoginIdentityDetailsErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	testcases := map[string]struct {
		certificateProviderStore func() *mockCertificateProviderStore
		template                 func() *mockTemplate
	}{
		"when certificateProviderStore error": {
			certificateProviderStore: func() *mockCertificateProviderStore {
				store := newMockCertificateProviderStore(t)
				store.EXPECT().
					Get(r.Context()).
					Return(nil, expectedError)
				return store
			},
			template: func() *mockTemplate { return newMockTemplate(t) },
		},
		"when template error": {
			certificateProviderStore: func() *mockCertificateProviderStore {
				store := newMockCertificateProviderStore(t)
				store.EXPECT().
					Get(mock.Anything).
					Return(&actor.CertificateProviderProvidedDetails{}, nil)
				return store
			},
			template: func() *mockTemplate {
				template := newMockTemplate(t)
				template.EXPECT().
					Execute(mock.Anything, mock.Anything).
					Return(expectedError)
				return template
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := OneLoginIdentityDetails(tc.template().Execute, tc.certificateProviderStore(), nil)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostOneLoginIdentityDetails(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	updatedCertificateProvider := &actor.CertificateProviderProvidedDetails{
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
		LpaID:            "lpa-id",
		DateOfBirth:      date.New("2000", "1", "1"),
		Tasks:            actor.CertificateProviderTasks{ConfirmYourIdentity: actor.TaskCompleted},
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
			DateOfBirth:      date.New("2000", "1", "1"),
			LpaID:            "lpa-id",
		}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), updatedCertificateProvider).
		Return(nil)

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpastore.CertificateProvider{FirstNames: "a", LastName: "b"}}, nil)

	err := OneLoginIdentityDetails(nil, certificateProviderStore, lpaResolvingService)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsWhenDetailsDoNotMatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed, FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1")},
			DateOfBirth:      date.New("2000", "1", "1"),
			LpaID:            "lpa-id",
		}, nil)

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{LpaUID: "lpa-uid", CertificateProvider: lpastore.CertificateProvider{FirstNames: "x", LastName: "y"}}, nil)

	err := OneLoginIdentityDetails(nil, certificateProviderStore, lpaResolvingService)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.ProveYourIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostOneLoginIdentityDetailsErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	ignoreLpaStoreClient := func() *mockLpaStoreClient { return newMockLpaStoreClient(t) }

	testcases := map[string]struct {
		lpaResolvingService      func() *mockLpaStoreResolvingService
		certificateProviderStore func() *mockCertificateProviderStore
		lpaStoreClient           func() *mockLpaStoreClient
	}{
		"when lpaStoreResolvingService error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(mock.Anything).
					Return(&lpastore.Lpa{}, expectedError)
				return service
			},
			certificateProviderStore: func() *mockCertificateProviderStore {
				store := newMockCertificateProviderStore(t)
				store.EXPECT().
					Get(mock.Anything).
					Return(&actor.CertificateProviderProvidedDetails{DateOfBirth: date.New("2000", "1", "1")}, nil)
				return store
			},
			lpaStoreClient: ignoreLpaStoreClient,
		},
		"when certificateProviderStore.Put() error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(mock.Anything).
					Return(&lpastore.Lpa{CertificateProvider: lpastore.CertificateProvider{FirstNames: "a", LastName: "b"}}, nil)
				return service
			},
			certificateProviderStore: func() *mockCertificateProviderStore {
				store := newMockCertificateProviderStore(t)
				store.EXPECT().
					Get(mock.Anything).
					Return(&actor.CertificateProviderProvidedDetails{
						IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", DateOfBirth: date.New("2000", "1", "1"), Status: identity.StatusConfirmed},
						DateOfBirth:      date.New("2000", "1", "1"),
					}, nil)
				store.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(expectedError)
				return store
			},
			lpaStoreClient: ignoreLpaStoreClient,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := OneLoginIdentityDetails(nil, tc.certificateProviderStore(), tc.lpaResolvingService())(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
