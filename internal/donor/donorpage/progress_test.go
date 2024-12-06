package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProgress(t *testing.T) {
	testCases := map[string]struct {
		provided          *donordata.Provided
		infoNotifications []progressNotification
	}{
		"none": {
			provided: &donordata.Provided{LpaUID: "lpa-uid"},
		},
		"going to the post office": {
			provided: &donordata.Provided{
				LpaUID: "lpa-uid",
				Tasks: donordata.Tasks{
					ConfirmYourIdentity: task.IdentityStatePending,
				},
			},
			infoNotifications: []progressNotification{
				{
					Heading: "youHaveChosenToConfirmYourIdentityAtPostOffice",
					Body:    "whenYouHaveConfirmedAtPostOfficeReturnToTaskList",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
				Return(task.Progress{DonorSigned: task.ProgressTask{Done: true}})

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{
					App:               testAppData,
					Donor:             tc.provided,
					Progress:          task.Progress{DonorSigned: task.ProgressTask{Done: true}},
					InfoNotifications: tc.infoNotifications,
				}).
				Return(nil)

			err := Progress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetProgressWhenLpaStoreClientErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := Progress(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}

func TestGetProgressOnTemplateError(t *testing.T) {
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

	err := Progress(template.Execute, lpaStoreResolvingService, progressTracker)(testAppData, w, r, &donordata.Provided{LpaUID: "lpa-uid"})
	assert.Equal(t, expectedError, err)
}
