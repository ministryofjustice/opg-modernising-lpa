package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.ResolvedLpa{}
	certificateProvider := &actor.CertificateProviderProvidedDetails{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(certificateProvider, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmYourDetailsData{App: testAppData, Lpa: donor, CertificateProvider: certificateProvider}).
		Return(nil)

	err := ConfirmYourDetails(template.Execute, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestConfirmYourDetailsWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := ConfirmYourDetails(nil, nil, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestConfirmYourDetailsWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.ResolvedLpa{}, expectedError)

	err := ConfirmYourDetails(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.ResolvedLpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	testCases := map[string]struct {
		signedAt time.Time
		redirect page.CertificateProviderPath
	}{
		"signed":     {signedAt: time.Now(), redirect: page.Paths.CertificateProvider.TaskList},
		"not signed": {redirect: page.Paths.CertificateProvider.YourRole},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpastore.ResolvedLpa{SignedAt: tc.signedAt}, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Get(r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
			certificateProviderStore.EXPECT().
				Put(r.Context(), &actor.CertificateProviderProvidedDetails{
					LpaID: "lpa-id",
					Tasks: actor.CertificateProviderTasks{
						ConfirmYourDetails: actor.TaskCompleted,
					},
				}).
				Return(nil)

			err := ConfirmYourDetails(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.ResolvedLpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Get(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, lpaStoreResolvingService, certificateProviderStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}
