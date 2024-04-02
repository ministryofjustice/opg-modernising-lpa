package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	lpastore "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	resolvedLpa := &lpastore.ResolvedLpa{LpaUID: "lpa-uid"}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(resolvedLpa, nil)

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
		Progress(resolvedLpa, &actor.CertificateProviderProvidedDetails{}, []*actor.AttorneyProvidedDetails{}).
		Return(page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaProgressData{
			App:      testAppData,
			Donor:    &actor.DonorProvidedDetails{LpaUID: "lpa-uid"},
			Progress: page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, lpaStoreClient, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.DonorProvidedDetails{LpaUID: "lpa-uid"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(nil, expectedError)

	err := LpaProgress(nil, lpaStoreClient, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(&lpastore.ResolvedLpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := LpaProgress(nil, lpaStoreClient, certificateProviderStore, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(&lpastore.ResolvedLpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return(nil, expectedError)

	err := LpaProgress(nil, lpaStoreClient, certificateProviderStore, attorneyStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(&lpastore.ResolvedLpa{}, nil)

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

	err := LpaProgress(template.Execute, lpaStoreClient, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
