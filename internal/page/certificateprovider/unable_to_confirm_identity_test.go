package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUnableToConfirmIdentity(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	lpaResolvingService := newMockLpaStoreResolvingService(t)
	lpaResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{Donor: lpastore.Donor{FirstNames: "a"}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &unableToConfirmIdentityData{
			App:   testAppData,
			Donor: lpastore.Donor{FirstNames: "a"},
		}).
		Return(nil)

	err := UnableToConfirmIdentity(template.Execute, nil, lpaResolvingService)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetUnableToConfirmIdentityErrors(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	testcases := map[string]struct {
		lpaResolvingService func() *mockLpaStoreResolvingService
		template            func() *mockTemplate
	}{
		"when lpaResolvingService error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(r.Context()).
					Return(nil, expectedError)
				return service
			},
			template: func() *mockTemplate { return newMockTemplate(t) },
		},
		"when template error": {
			lpaResolvingService: func() *mockLpaStoreResolvingService {
				service := newMockLpaStoreResolvingService(t)
				service.EXPECT().
					Get(r.Context()).
					Return(&lpastore.Lpa{}, nil)
				return service
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
			err := UnableToConfirmIdentity(tc.template().Execute, nil, tc.lpaResolvingService())(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestPostUnableToConfirmIdentity(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &actor.CertificateProviderProvidedDetails{
			LpaID: "lpa-id",
			Tasks: actor.CertificateProviderTasks{ConfirmYourIdentity: actor.TaskCompleted},
		}).
		Return(nil)

	err := UnableToConfirmIdentity(nil, certificateProviderStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.ReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostUnableToConfirmIdentityErrors(t *testing.T) {
	testcases := map[string]func() *mockCertificateProviderStore{
		"when certificateProviderStore.Get() error": func() *mockCertificateProviderStore {
			store := newMockCertificateProviderStore(t)
			store.EXPECT().
				Get(mock.Anything).
				Return(nil, expectedError)
			return store
		},
		"when certificateProviderStore.Put() error": func() *mockCertificateProviderStore {
			store := newMockCertificateProviderStore(t)
			store.EXPECT().
				Get(mock.Anything).
				Return(&actor.CertificateProviderProvidedDetails{}, nil)
			store.EXPECT().
				Put(mock.Anything, mock.Anything).
				Return(expectedError)
			return store
		},
	}

	for name, certificateProviderStore := range testcases {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			w := httptest.NewRecorder()

			err := UnableToConfirmIdentity(nil, certificateProviderStore(), nil)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}