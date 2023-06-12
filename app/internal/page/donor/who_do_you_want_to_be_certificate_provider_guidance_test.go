package donor

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhoDoYouWantToBeCertificateProviderGuidance(t *testing.T) {
	testCases := map[string]struct {
		data       *page.Lpa
		notStarted bool
	}{
		"unset": {
			data:       &page.Lpa{},
			notStarted: true,
		},
		"in-progress": {
			data:       &page.Lpa{Tasks: page.Tasks{CertificateProvider: actor.TaskInProgress}},
			notStarted: false,
		},
		"completed": {
			data:       &page.Lpa{Tasks: page.Tasks{CertificateProvider: actor.TaskCompleted}},
			notStarted: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &whoDoYouWantToBeCertificateProviderGuidanceData{
					App: testAppData,
					Lpa: tc.data,
				}).
				Return(nil)

			err := WhoDoYouWantToBeCertificateProviderGuidance(template.Execute, nil)(testAppData, w, r, tc.data)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetWhoDoYouWantToBeCertificateProviderGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{Tasks: page.Tasks{CertificateProvider: actor.TaskInProgress}}).
		Return(nil)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, donorStore)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.CertificateProviderDetails, resp.Header.Get("Location"))
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidanceWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}
