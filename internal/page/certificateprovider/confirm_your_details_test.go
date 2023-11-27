package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{}
	certificateProvider := &actor.CertificateProviderProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(certificateProvider, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &confirmYourDetailsData{App: testAppData, Donor: donor, CertificateProvider: certificateProvider}).
		Return(nil)

	err := ConfirmYourDetails(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestConfirmYourDetailsWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := ConfirmYourDetails(nil, nil, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestConfirmYourDetailsWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, expectedError)

	err := ConfirmYourDetails(nil, donorStore, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, donorStore, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{
			LpaID: "lpa-id",
			Tasks: actor.CertificateProviderTasks{
				ConfirmYourDetails: actor.TaskCompleted,
			},
		}).
		Return(nil)

	err := ConfirmYourDetails(nil, donorStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProvider.YourRole.Format("lpa-id"), resp.Header.Get("Location"))

}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.DonorProvidedDetails{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, donorStore, certificateProviderStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}
