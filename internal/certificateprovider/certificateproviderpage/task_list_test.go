package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa                 *lpadata.Lpa
		certificateProvider *certificateproviderdata.Provided
		appData             appcontext.Data
		expected            func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa:                 &lpadata.Lpa{LpaID: "lpa-id"},
			certificateProvider: &certificateproviderdata.Provided{},
			appData:             testAppData,
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"submitted": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			},
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = certificateprovider.PathConfirmYourDetails

				return items
			},
		},
		"identity confirmation in progress": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Paid:                             true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.IdentityStateInProgress,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = certificateprovider.PathConfirmYourDetails
				items[1].IdentityState = task.IdentityStateInProgress
				items[1].Path = certificateprovider.PathHowWillYouConfirmYourIdentity

				return items
			},
		},
		"identity confirmed": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Paid:                             true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:    task.StateCompleted,
					ConfirmYourIdentity:   task.IdentityStateCompleted,
					ProvideTheCertificate: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = certificateprovider.PathConfirmYourDetails
				items[1].IdentityState = task.IdentityStateCompleted
				items[1].Path = certificateprovider.PathReadTheLpa
				items[2].State = task.StateCompleted

				return items
			},
		},
		"all": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Paid:                             true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:    task.StateCompleted,
					ConfirmYourIdentity:   task.IdentityStateCompleted,
					ProvideTheCertificate: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = certificateprovider.PathConfirmYourDetails
				items[1].IdentityState = task.IdentityStateCompleted
				items[1].Path = certificateprovider.PathReadTheLpa
				items[2].State = task.StateCompleted

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App:      tc.appData,
					Provided: tc.certificateProvider,
					Lpa:      tc.lpa,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: certificateprovider.PathEnterDateOfBirth},
						{Name: "confirmYourIdentity", Path: certificateprovider.PathConfirmYourIdentity},
						{Name: "provideYourCertificate", Path: certificateprovider.PathReadTheLpa},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute)(tc.appData, w, r, tc.certificateProvider, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
