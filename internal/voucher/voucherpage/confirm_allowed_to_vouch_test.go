package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmAllowedToVouch(t *testing.T) {
	testcases := map[string]struct {
		lpa      *lpadata.Lpa
		provided *voucherdata.Provided
		data     *confirmAllowedToVouchData
	}{
		"actor matches": {
			lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			provided: &voucherdata.Provided{
				LpaID:      "lpa-id",
				FirstNames: "V",
				LastName:   "W",
			},
			data: &confirmAllowedToVouchData{
				App:  testAppData,
				Form: form.NewYesNoForm(form.YesNoUnknown),
				Lpa: &lpadata.Lpa{
					Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
				},
			},
		},
		"surname matches donor": {
			lpa: &lpadata.Lpa{
				Donor:   lpadata.Donor{FirstNames: "A", LastName: "W"},
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			provided: &voucherdata.Provided{
				LpaID:      "lpa-id",
				FirstNames: "V",
				LastName:   "W",
			},
			data: &confirmAllowedToVouchData{
				App:  testAppData,
				Form: form.NewYesNoForm(form.YesNoUnknown),
				Lpa: &lpadata.Lpa{
					Donor:   lpadata.Donor{FirstNames: "A", LastName: "W"},
					Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
				},
				SurnameMatchesDonor: true,
			},
		},
		"matches actor after identity": {
			lpa: &lpadata.Lpa{
				Donor:   lpadata.Donor{FirstNames: "A", LastName: "W"},
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			provided: &voucherdata.Provided{
				LpaID:      "lpa-id",
				FirstNames: "V",
				LastName:   "W",
				Tasks: voucherdata.Tasks{
					ConfirmYourIdentity: task.StateInProgress,
				},
			},
			data: &confirmAllowedToVouchData{
				App:  testAppData,
				Form: form.NewYesNoForm(form.YesNoUnknown),
				Lpa: &lpadata.Lpa{
					Donor:   lpadata.Donor{FirstNames: "A", LastName: "W"},
					Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
				},
				SurnameMatchesDonor: true,
				MatchIdentity:       true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.data).
				Return(nil)

			err := ConfirmAllowedToVouch(template.Execute, lpaStoreResolvingService, nil, nil)(testAppData, w, r, tc.provided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetConfirmAllowedToVouchWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, nil, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmAllowedToVouchWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmAllowedToVouch(template.Execute, lpaStoreResolvingService, nil, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmAllowedToVouch(t *testing.T) {
	testcases := map[task.State]voucherdata.Tasks{
		task.StateNotStarted: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
		task.StateInProgress: voucherdata.Tasks{ConfirmYourIdentity: task.StateCompleted},
	}

	for taskState, tasks := range testcases {
		t.Run(taskState.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.Yes.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

			voucherStore := newMockVoucherStore(t)
			voucherStore.EXPECT().
				Put(r.Context(), &voucherdata.Provided{LpaID: "lpa-id", Tasks: tasks}).
				Return(nil)

			err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id", Tasks: voucherdata.Tasks{ConfirmYourIdentity: taskState}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostConfirmAllowedToVouchWhenNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &voucherdata.Provided{LpaID: "lpa-id", FirstNames: "a", LastName: "b"}
	lpa := &lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith", Email: "a@example.com"}, LpaUID: "lpa-uid"}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	failVouch := newMockFailVouch(t)
	failVouch.EXPECT().
		Execute(r.Context(), provided, lpa).
		Return(nil)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, nil, failVouch.Execute)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathYouCannotVouchForDonor.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmAllowedToVouchWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{}, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostConfirmAllowedToVouchWhenNoWhenFailVouchError(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	failVouch := newMockFailVouch(t)
	failVouch.EXPECT().
		Execute(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, nil, failVouch.Execute)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
