package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/stretchr/testify/assert"
)

func TestGetChooseNewCertificateProvider(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseNewCertificateProviderData{Donor: &actor.DonorProvidedDetails{}, App: testAppData}).
		Return(nil)

	err := ChooseNewCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseNewCertificateProviderWhenTemplateError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseNewCertificateProviderData{Donor: &actor.DonorProvidedDetails{}, App: testAppData}).
		Return(expectedError)

	err := ChooseNewCertificateProvider(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseNewCertificateProvider(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{}}).
		Return(nil)

	err := ChooseNewCertificateProvider(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{FirstNames: "first-names"}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestPostChooseNewCertificateProviderWhenStoreError(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{}}).
		Return(expectedError)

	err := ChooseNewCertificateProvider(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{CertificateProvider: donordata.CertificateProvider{FirstNames: "first-names"}})

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
