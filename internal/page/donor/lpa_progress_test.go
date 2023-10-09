package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAny", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaProgressData{
			App:      testAppData,
			Lpa:      &page.Lpa{ID: "123"},
			Progress: page.Progress{DonorSigned: actor.TaskInProgress},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, nil)(testAppData, w, r, &page.Lpa{ID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAny", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, attorneyStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("GetAny", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAny", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore)(testAppData, w, r, &page.Lpa{ID: "123"})
	assert.Equal(t, expectedError, err)
}
