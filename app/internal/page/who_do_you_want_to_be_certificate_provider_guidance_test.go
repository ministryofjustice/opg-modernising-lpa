package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhoDoYouWantToBeCertificateProviderGuidance(t *testing.T) {
	testCases := map[string]struct {
		data       *Lpa
		notStarted bool
	}{
		"unset": {
			data:       &Lpa{},
			notStarted: true,
		},
		"in-progress": {
			data:       &Lpa{Tasks: Tasks{CertificateProvider: TaskInProgress}},
			notStarted: false,
		},
		"completed": {
			data:       &Lpa{Tasks: Tasks{CertificateProvider: TaskCompleted}},
			notStarted: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(tc.data, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, &whoDoYouWantToBeCertificateProviderGuidanceData{
					App:        appData,
					NotStarted: tc.notStarted,
					Lpa:        tc.data,
				}).
				Return(nil)

			err := WhoDoYouWantToBeCertificateProviderGuidance(template.Func, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, template, lpaStore)
		})
	}
}

func TestGetWhoDoYouWantToBeCertificateProviderGuidanceWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWhoDoYouWantToBeCertificateProviderGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Tasks: Tasks{CertificateProvider: TaskInProgress}}).
		Return(nil)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.CertificateProviderDetails, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidanceWhenWillDoLater(t *testing.T) {
	form := url.Values{"will-do-this-later": {"1"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.TaskList, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhoDoYouWantToBeCertificateProviderGuidanceWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WhoDoYouWantToBeCertificateProviderGuidance(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}
