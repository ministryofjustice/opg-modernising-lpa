package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderStart(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?sessionId=123&lpaId=456", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := sessionDataFromContext(ctx)

			return assert.Equal(t, &sessionData{SessionID: "123", LpaID: "456"}, session)
		})).
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &certificateProviderStartData{
			App:   appData,
			Start: Paths.CertificateProviderLogin + "?lpaId=456&sessionId=123",
		}).
		Return(nil)

	err := CertificateProviderStart(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestCertificateProviderStartWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(lpa, expectedError)

	err := CertificateProviderStart(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestCertificateProviderStartWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", mock.Anything, mock.Anything).
		Return(expectedError)

	err := CertificateProviderStart(template.Func, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
