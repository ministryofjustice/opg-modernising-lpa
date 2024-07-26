package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/stretchr/testify/assert"
)

func TestGuidance(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{}
	certificateProvider := &actor.CertificateProviderProvidedDetails{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Lpa: donor, CertificateProvider: certificateProvider}).
		Return(nil)

	err := Guidance(template.Execute, lpaStoreResolvingService)(testAppData, w, r, certificateProvider)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenNilDataStores(t *testing.T) {
	w := httptest.NewRecorder()

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Guidance(template.Execute, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGuidanceWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, expectedError)

	err := Guidance(nil, lpaStoreResolvingService)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGuidanceWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &guidanceData{App: testAppData, Lpa: &lpastore.Lpa{}}).
		Return(expectedError)

	err := Guidance(template.Execute, lpaStoreResolvingService)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}
