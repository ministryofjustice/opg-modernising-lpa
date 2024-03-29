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
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(&actor.DonorProvidedDetails{LpaID: "123"}, &actor.CertificateProviderProvidedDetails{}, []*actor.AttorneyProvidedDetails{}).
		Return(page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaProgressData{
			App:      testAppData,
			Donor:    &actor.DonorProvidedDetails{LpaID: "123"},
			Progress: page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, attorneyStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything, mock.Anything, mock.Anything).
		Return(page.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
