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

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
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

	lpa := &lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}
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

	vouchFailer := newMockVouchFailer(t)
	vouchFailer.EXPECT().
		Execute(r.Context(), provided, lpa).
		Return(nil)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, vouchFailer.Execute)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathDonorDetailsDoNotMatch.Format("lpa-id"), resp.Header.Get("Location"))
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

func TestPostVerifyDonorDetailsWhenFailVouchErrors(t *testing.T) {
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

	vouchFailer := newMockVouchFailer(t)
	vouchFailer.EXPECT().
		Execute(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := VerifyDonorDetails(nil, lpaStoreResolvingService, voucherStore, vouchFailer.Execute)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
