package voucherpage

import (
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

	err := VerifyDonorDetails(template.Execute, lpaStoreResolvingService, nil, nil, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
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

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, nil, nil, nil)(testAppData, w, r, nil)

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

	err := VerifyDonorDetails(template.Execute, lpaStoreResolvingService, nil, nil, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostVerifyDonorDetailsWhenYes(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
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
			DonorDetailsMatch: form.Yes,
			Tasks:             voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{VouchAttempts: 1, DetailsVerifiedByVoucher: true}).
		Return(nil)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, nil, donorStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostVerifyDonorDetailsWhenNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	w := httptest.NewRecorder()
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &lpadata.Lpa{Donor: lpadata.Donor{FirstNames: "John", LastName: "Smith"}}
	provided := &voucherdata.Provided{
		LpaID:             "lpa-id",
		DonorDetailsMatch: form.No,
		Tasks:             voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
	}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(lpa, nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), provided).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{VouchAttempts: 1}).
		Return(nil)

	vouchFailer := newMockVouchFailer(t)
	vouchFailer.EXPECT().
		Execute(r.Context(), provided, lpa).
		Return(nil)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, vouchFailer.Execute, donorStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathVoucherDonorDetailsDoNotMatch.Format()+"?donorFirstNames=John&donorFullName=John+Smith", resp.Header.Get("Location"))
}

func TestPostVerifyDonorDetailsWhenVoucherStoreErrors(t *testing.T) {
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

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, nil, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostVerifyDonorDetailsWhenDonorStoreErrors(t *testing.T) {
	testcases := map[string]struct {
		setupDonorStore func(*testing.T) *mockDonorStore
	}{
		"GetAny": {
			setupDonorStore: func(t *testing.T) *mockDonorStore {
				s := newMockDonorStore(t)
				s.EXPECT().
					GetAny(mock.Anything).
					Return(&donordata.Provided{}, expectedError)
				return s
			},
		},
		"Put": {
			setupDonorStore: func(t *testing.T) *mockDonorStore {
				s := newMockDonorStore(t)
				s.EXPECT().
					GetAny(mock.Anything).
					Return(&donordata.Provided{}, nil)
				s.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(expectedError)
				return s
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {form.Yes.String()},
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
					DonorDetailsMatch: form.Yes,
					Tasks:             voucherdata.Tasks{VerifyDonorDetails: task.StateCompleted},
				}).
				Return(nil)

			err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, nil, tc.setupDonorStore(t))(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})

			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestPostVerifyDonorDetailsWhenFailVouchErrors(t *testing.T) {
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

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(mock.Anything).
		Return(&donordata.Provided{}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	vouchFailer := newMockVouchFailer(t)
	vouchFailer.EXPECT().
		Execute(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, vouchFailer.Execute, donorStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.ErrorIs(t, err, expectedError)
}
