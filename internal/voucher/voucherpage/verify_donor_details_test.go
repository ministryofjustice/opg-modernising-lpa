package voucherpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVerifyDonorDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
		}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &verifyDonorDetailsData{
			App: testAppData,
			Lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := VerifyDonorDetails(template.Execute, lpaStoreResolvingService, nil, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetVerifyDonorDetailsWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, nil, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetVerifyDonorDetailsWhenTemplateErrors(t *testing.T) {
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

	err := VerifyDonorDetails(template.Execute, lpaStoreResolvingService, nil, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostVerifyDonorDetails(t *testing.T) {

	testcases := map[form.YesNo]struct {
		expectedRedirect voucher.Path
		donorStore       func() *mockDonorStore
	}{
		form.Yes: {
			expectedRedirect: voucher.PathTaskList,
			donorStore:       func() *mockDonorStore { return newMockDonorStore(t) },
		},
		form.No: {
			expectedRedirect: voucher.PathDonorDetailsDoNotMatch,
			donorStore: func() *mockDonorStore {
				d := newMockDonorStore(t)
				d.EXPECT().
					GetAny(context.Background()).
					Return(&donordata.Provided{}, nil)
				d.EXPECT().
					Put(context.Background(), &donordata.Provided{FailedVouchAttempts: 1}).
					Return(nil)
				return d
			},
		},
	}

	for yesNo, tc := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {yesNo.String()},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			w := httptest.NewRecorder()
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

			voucherStore := newMockVoucherStore(t)
			voucherStore.EXPECT().
				Put(r.Context(), &voucherdata.Provided{
					LpaID:             "lpa-id",
					DonorDetailsMatch: yesNo,
					Tasks:             voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
				}).
				Return(nil)

			err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, tc.donorStore())(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostVerifyDonorDetailsWhenStoreErrors(t *testing.T) {
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

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}

func TestPostVerifyDonorDetailsWhenDonorStoreErrors(t *testing.T) {
	testcases := map[string]func() *mockDonorStore{
		"GetAny": func() *mockDonorStore {
			d := newMockDonorStore(t)
			d.EXPECT().
				GetAny(mock.Anything).
				Return(&donordata.Provided{}, expectedError)
			return d
		},
		"Put": func() *mockDonorStore {
			d := newMockDonorStore(t)
			d.EXPECT().
				GetAny(mock.Anything).
				Return(&donordata.Provided{}, nil)
			d.EXPECT().
				Put(mock.Anything, mock.Anything).
				Return(expectedError)
			return d
		},
	}

	for name, donorStore := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.No.String()},
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
				Return(nil)

			err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, donorStore())(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
			assert.Equal(t, expectedError, err)
		})
	}

}
