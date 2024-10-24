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
		donor               *lpadata.Lpa
		certificateProvider *certificateproviderdata.Provided
		appData             appcontext.Data
		expected            func([]taskListItem) []taskListItem
	}{
		"empty": {
			donor:               &lpadata.Lpa{LpaID: "lpa-id"},
			certificateProvider: &certificateproviderdata.Provided{},
			appData:             testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"paid": {
			donor: &lpadata.Lpa{
				LpaID: "lpa-id",
				Paid:  true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"submitted": {
			donor: &lpadata.Lpa{
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
				items[1].Disabled = true
				items[2].Disabled = true

				return items
			},
		},
		"identity confirmed": {
			donor: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Paid:                             true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:    task.StateCompleted,
					ConfirmYourIdentity:   task.StateCompleted,
					ProvideTheCertificate: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[1].Path = certificateprovider.PathReadTheLpa.Format("lpa-id")
				items[2].State = task.StateCompleted

				return items
			},
		},
		"all": {
			donor: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Paid:                             true,
			},
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:    task.StateCompleted,
					ConfirmYourIdentity:   task.StateCompleted,
					ProvideTheCertificate: task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[1].Path = certificateprovider.PathReadTheLpa.Format("lpa-id")
				items[2].State = task.StateCompleted

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.donor, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App: tc.appData,
					Lpa: tc.donor,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: certificateprovider.PathEnterDateOfBirth.Format("lpa-id")},
						{Name: "confirmYourIdentity", Path: certificateprovider.PathConfirmYourIdentity.Format("lpa-id")},
						{Name: "provideYourCertificate", Path: certificateprovider.PathReadTheLpa.Format("lpa-id")},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStoreResolvingService)(tc.appData, w, r, tc.certificateProvider)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, expectedError)

	err := TaskList(nil, lpaStoreResolvingService)(testAppData, w, r, &certificateproviderdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaID: "lpa-id"}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &certificateproviderdata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
