package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{}
	certificateProvider := &actor.CertificateProviderProvidedDetails{}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(certificateProvider, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: testAppData, Lpa: lpa, CertificateProvider: certificateProvider}).
		Return(nil)

	err := Guidance(template.Execute, lpaStore, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenNilDataStores(t *testing.T) {
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: testAppData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, expectedError)

	err := Guidance(nil, lpaStore, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGuidanceWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := Guidance(nil, nil, certificateProviderStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &guidanceData{App: testAppData, Lpa: &page.Lpa{}}).
		Return(expectedError)

	err := Guidance(template.Execute, lpaStore, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}
