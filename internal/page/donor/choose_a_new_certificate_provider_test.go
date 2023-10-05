package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetChooseNewCertificateProvider(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseNewCertificateProviderData{Lpa: &page.Lpa{}, App: testAppData}).
		Return(nil)

	err := ChooseNewCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseNewCertificateProviderWhenTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseNewCertificateProviderData{Lpa: &page.Lpa{}, App: testAppData}).
		Return(expectedError)

	err := ChooseNewCertificateProvider(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseNewCertificateProvider(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{CertificateProvider: actor.CertificateProvider{}}).
		Return(nil)

	err := ChooseNewCertificateProvider(nil, donorStore)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{FirstNames: "first-names"}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestPostChooseNewCertificateProviderWhenStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{CertificateProvider: actor.CertificateProvider{}}).
		Return(expectedError)

	err := ChooseNewCertificateProvider(nil, donorStore)(testAppData, w, r, &page.Lpa{CertificateProvider: actor.CertificateProvider{FirstNames: "first-names"}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
