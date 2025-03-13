package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	testCases := map[string]struct {
		lpa      *lpadata.Lpa
		voucher  *voucherdata.Provided
		expected func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"},
			},
			voucher: &voucherdata.Provided{},
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"identity confirmation in progress": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"},
			},
			voucher: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:     task.StateCompleted,
					VerifyDonorDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.IdentityStateInProgress,
				},
			},
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[2].IdentityState = task.IdentityStateInProgress
				items[2].Path = voucher.PathHowWillYouConfirmYourIdentity
				return items
			},
		},
		"identity checked but matches actor": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"},
			},
			voucher: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:     task.StateCompleted,
					VerifyDonorDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.IdentityStateInProgress,
				},
				IdentityUserData: identity.UserData{CheckedAt: time.Now()},
			},
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[2].IdentityState = task.IdentityStateInProgress
				items[2].Path = voucher.PathConfirmAllowedToVouch
				return items
			},
		},
		"completed": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"},
			},
			voucher: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:     task.StateCompleted,
					VerifyDonorDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.IdentityStateCompleted,
					SignTheDeclaration:  task.StateCompleted,
				},
			},
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[2].IdentityState = task.IdentityStateCompleted
				items[2].Path = voucher.PathOneLoginIdentityDetails
				items[3].State = task.StateCompleted
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
				Return(tc.lpa, nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				Format("verifyPersonIdentity", map[string]any{"DonorFullName": "John Smith"}).
				Return("verifyJohnSmithsIdentity")

			appData := testAppData
			appData.Localizer = localizer

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App:     appData,
					Voucher: tc.voucher,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourName", Path: voucher.PathConfirmYourName},
						{Name: "verifyJohnSmithsIdentity", Path: voucher.PathVerifyDonorDetails},
						{Name: "confirmYourIdentity", Path: voucher.PathConfirmYourIdentity},
						{Name: "signTheDeclaration", Path: voucher.PathSignTheDeclaration},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute, lpaStoreResolvingService)(appData, w, r, tc.voucher)
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

	err := TaskList(nil, lpaStoreResolvingService)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaID: "lpa-id"}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		Format(mock.Anything, mock.Anything).
		Return("hey")

	appData := testAppData
	appData.Localizer = localizer

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute, lpaStoreResolvingService)(appData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
