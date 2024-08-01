package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	lpastore "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpastore.Lpa{LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(lpa).
		Return(page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaProgressData{
			App:      testAppData,
			Donor:    &donordata.DonorProvidedDetails{LpaUID: "lpa-uid"},
			Progress: page.Progress{DonorSigned: page.ProgressTask{State: actor.TaskInProgress}},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaUID: "lpa-uid"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := LpaProgress(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything).
		Return(page.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LpaProgress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
