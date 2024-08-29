package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notification"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(lpa, donordata.Tasks{YourDetails: task.StateCompleted}, notification.Notifications{FeeEvidence: notification.Notification{Received: testNow}}, pay.FullFee).
		Return(task.Progress{DonorSigned: task.ProgressTask{State: task.StateInProgress}})

	donor := &donordata.Provided{
		LpaUID:        "lpa-uid",
		Tasks:         donordata.Tasks{YourDetails: task.StateCompleted},
		Notifications: notification.Notifications{FeeEvidence: notification.Notification{Received: testNow}},
		FeeType:       pay.FullFee,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaProgressData{
			App:      testAppData,
			Donor:    donor,
			Progress: task.Progress{DonorSigned: task.ProgressTask{State: task.StateInProgress}},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, donor)
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

	err := LpaProgress(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	progressTracker := newMockProgressTracker(t)
	progressTracker.EXPECT().
		Progress(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(task.Progress{})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LpaProgress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
