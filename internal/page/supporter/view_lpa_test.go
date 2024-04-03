package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	lpastore "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetViewLPA(t *testing.T) {
	testcases := map[string]error{
		"with actors":    nil,
		"without actors": dynamo.NotFoundError{},
	}

	for name, storeError := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", LpaUID: "lpa-uid"}
			resolvedLpa := &lpastore.Lpa{LpaUID: "lpa-uid"}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(r.Context()).
				Return(donor, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(resolvedLpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(r.Context()).
				Return(&actor.CertificateProviderProvidedDetails{}, storeError)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				GetAny(r.Context()).
				Return([]*actor.AttorneyProvidedDetails{{}}, storeError)

			progressTracker := newMockProgressTracker(t)
			progressTracker.EXPECT().
				Progress(resolvedLpa, &actor.CertificateProviderProvidedDetails{}, []*actor.AttorneyProvidedDetails{{}}).
				Return(page.Progress{Paid: page.ProgressTask{State: actor.TaskInProgress}})

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &viewLPAData{
					App:      testAppData,
					Donor:    donor,
					Progress: page.Progress{Paid: page.ProgressTask{State: actor.TaskInProgress}},
				}).
				Return(nil)

			err := ViewLPA(template.Execute, lpaStoreResolvingService, donorStore, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.Organisation{}, nil)

			assert.Nil(t, err)
		})
	}

}

func TestGetViewLPAWithDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, nil, donorStore, nil, nil, nil)(testAppData, w, r, &actor.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenLpaStoreClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, donorStore, nil, nil, nil)(testAppData, w, r, &actor.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, donorStore, certificateProviderStore, nil, nil)(testAppData, w, r, &actor.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenAttorneyStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, donorStore, certificateProviderStore, attorneyStore, nil)(testAppData, w, r, &actor.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(mock.Anything).
		Return([]*actor.AttorneyProvidedDetails{{}}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything, mock.Anything, mock.Anything).
		Return(page.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ViewLPA(template.Execute, lpaStoreResolvingService, donorStore, certificateProviderStore, attorneyStore, progressTracker)(testAppData, w, r, &actor.Organisation{}, nil)

	assert.Error(t, err)
}
