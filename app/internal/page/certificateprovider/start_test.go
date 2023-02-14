package certificateprovider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCertificateProviderStart(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?sessionId=123&lpaId=456", nil)

	lpa := &page.Lpa{}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "123", LpaID: "456"}, session)
		})).
		Return(lpa, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &startData{
			App:   page.TestAppData,
			Start: page.Paths.CertificateProviderLogin + "?lpaId=456&sessionId=123",
		}).
		Return(nil)

	err := Start(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestCertificateProviderStartWhenDataStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(lpa, page.ExpectedError)

	err := Start(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestCertificateProviderStartWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", mock.Anything, mock.Anything).
		Return(page.ExpectedError)

	err := Start(template.Func, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
