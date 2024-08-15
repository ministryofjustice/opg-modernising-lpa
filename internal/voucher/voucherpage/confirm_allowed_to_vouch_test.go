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
		Execute(w, &confirmAllowedToVouchData{
			App: testAppData,
			Lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := ConfirmAllowedToVouch(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmAllowedToVouchWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)

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

	err := ConfirmAllowedToVouch(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmAllowedToVouch(t *testing.T) {
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
		Put(r.Context(), &voucherdata.Provided{
			LpaID: "lpa-id",
			Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
		}).
		Return(nil)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmAllowedToVouchWhenNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{LastName: "Smith"}}, nil)

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Error(t, err)
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

	err := ConfirmAllowedToVouch(nil, lpaStoreResolvingService, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
