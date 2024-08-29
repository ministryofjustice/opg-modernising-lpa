package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notification"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
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
		Progress(
			lpa,
			donordata.Tasks{YourDetails: task.StateCompleted},
			notification.Notifications{FeeEvidence: notification.Notification{Received: testNow}},
			pay.FullFee).
		Return(task.Progress{Paid: task.ProgressTask{State: task.StateInProgress}})

	donor := &donordata.Provided{
		LpaUID:        "lpa-uid",
		Tasks:         donordata.Tasks{YourDetails: task.StateCompleted},
		Notifications: notification.Notifications{FeeEvidence: notification.Notification{Received: testNow}},
		FeeType:       pay.FullFee,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLPAData{
			App:      testAppData,
			Lpa:      lpa,
			Progress: task.Progress{Paid: task.ProgressTask{State: task.StateInProgress}},
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(r.Context()).
		Return(donor, nil)

	err := ViewLPA(template.Execute, lpaStoreResolvingService, progressTracker, donorStore)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Nil(t, err)
}

func TestGetViewLPAWhenLpaStoreClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, nil, nil)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&donordata.Provided{}, expectedError)

	err := ViewLPA(nil, lpaStoreResolvingService, nil, donorStore)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Error(t, err)
}

func TestGetViewLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(&donordata.Provided{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(task.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ViewLPA(template.Execute, lpaStoreResolvingService, progressTracker, donorStore)(testAppData, w, r, &supporterdata.Organisation{}, nil)

	assert.Error(t, err)
}
