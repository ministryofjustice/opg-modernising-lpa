package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetViewLPA(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(lpa).
		Return(task.Progress{Paid: task.ProgressTask{State: task.StateInProgress}})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLPAData{
			App:      testAppData,
			Lpa:      lpa,
			Progress: task.Progress{Paid: task.ProgressTask{State: task.StateInProgress}},
		}).
		Return(nil)

	err := ViewLPA(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Nil(t, err)
}

func TestGetViewLPAWhenLpaStoreClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything).
		Return(task.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ViewLPA(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Error(t, err)
}
