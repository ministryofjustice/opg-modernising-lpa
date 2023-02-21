package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
			data:       &page.Lpa{Tasks: page.Tasks{CertificateProvider: page.TaskInProgress}},
			notStarted: false,
		},
		"completed": {
			data:       &page.Lpa{Tasks: page.Tasks{CertificateProvider: page.TaskCompleted}},
			notStarted: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.data, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &whoDoYouWantToBeCertificateProviderGuidanceData{
					App:        testAppData,
					NotStarted: tc.notStarted,
					Lpa:        tc.data,
				}).
				Return(nil)

			err := WhoDoYouWantToBeCertificateProviderGuidance(template.Execute, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

		})
	}
}

func TestGetWhoDoYouWantToBeCertificateProviderGuidanceWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhoDoYouWantToBeCertificateProviderGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{Tasks: page.Tasks{CertificateProvider: page.TaskInProgress}}).
		Return(nil)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.CertificateProviderDetails, resp.Header.Get("Location"))
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidanceWhenWillDoLater(t *testing.T) {
	form := url.Values{"will-do-this-later": {"1"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidanceWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}
