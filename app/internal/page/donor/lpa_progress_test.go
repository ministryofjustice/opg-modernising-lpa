package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})

	certificateProviderStore.
		On("GetAny", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaProgressData{
			App:                 testAppData,
			Lpa:                 &page.Lpa{ID: "123"},
			CertificateProvider: &actor.CertificateProviderProvidedDetails{},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, certificateProviderStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})

	certificateProviderStore.
		On("GetAny", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "123"})

	certificateProviderStore.
		On("GetAny", ctx).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaProgressData{
			App:                 testAppData,
			Lpa:                 &page.Lpa{ID: "123"},
			CertificateProvider: &actor.CertificateProviderProvidedDetails{},
		}).
		Return(expectedError)

	err := LpaProgress(template.Execute, certificateProviderStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
