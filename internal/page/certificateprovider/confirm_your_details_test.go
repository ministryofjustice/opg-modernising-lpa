package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &actor.DonorProvidedDetails{}
	certificateProvider := &actor.CertificateProviderProvidedDetails{}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(certificateProvider, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &confirmYourDetailsData{App: testAppData, Donor: lpa, CertificateProvider: certificateProvider}).
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
	testCases := map[string]struct {
		confirmYourDetailsTaskState actor.TaskState
		expectedRedirect            string
		lpaSubmittedDate            time.Time
	}{
		"CP has not confirmed details": {
			confirmYourDetailsTaskState: actor.TaskNotStarted,
			expectedRedirect:            page.Paths.CertificateProvider.YourRole.Format("lpa-id"),
		},
		"CP has started to confirm details": {
			confirmYourDetailsTaskState: actor.TaskInProgress,
			expectedRedirect:            page.Paths.CertificateProvider.YourRole.Format("lpa-id"),
		},
		"CP has confirmed details": {
			confirmYourDetailsTaskState: actor.TaskCompleted,
			expectedRedirect:            page.Paths.CertificateProvider.TaskList.Format("lpa-id"),
		},
		"CP has witness donor sign": {
			confirmYourDetailsTaskState: actor.TaskNotStarted,
			expectedRedirect:            page.Paths.CertificateProvider.TaskList.Format("lpa-id"),
			lpaSubmittedDate:            time.Now(),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(&actor.DonorProvidedDetails{SignedAt: tc.lpaSubmittedDate}, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("Get", r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{LpaID: "lpa-id", Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: tc.confirmYourDetailsTaskState}}, nil)
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
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
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
