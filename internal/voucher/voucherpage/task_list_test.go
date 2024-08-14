package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		"completed": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"},
			},
			voucher: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:     task.StateCompleted,
					VerifyDonorDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.StateCompleted,
					SignTheDeclaration:  task.StateCompleted,
				},
			},
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[1].State = task.StateCompleted
				items[2].State = task.StateCompleted
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
				Possessive("John Smith").
				Return("John Smith's")
			localizer.EXPECT().
				Format("verifyDonorDetails", map[string]any{"DonorFullNamePossessive": "John Smith's"}).
				Return("verifyJohnSmithsDetails")

			appData := testAppData
			appData.Localizer = localizer

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App:     appData,
					Voucher: tc.voucher,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourName", Path: voucher.PathConfirmYourName.Format("lpa-id")},
						{Name: "verifyJohnSmithsDetails", Path: voucher.PathVerifyDonorDetails.Format("lpa-id")},
						{Name: "confirmYourIdentity", Path: voucher.PathConfirmYourIdentity.Format("lpa-id")},
						{Name: "signTheDeclaration", Path: voucher.PathSignTheDeclaration.Format("lpa-id")},
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
		Possessive(mock.Anything).
		Return("oi")
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
