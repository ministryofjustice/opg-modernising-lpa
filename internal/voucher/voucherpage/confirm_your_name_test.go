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

func TestGetConfirmYourName(t *testing.T) {
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
		Execute(w, &confirmYourNameData{
			App: testAppData,
			Lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			FirstNames: "V",
			LastName:   "W",
		}).
		Return(nil)

	err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourNameWhenChanged(t *testing.T) {
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
		Execute(w, &confirmYourNameData{
			App: testAppData,
			Lpa: &lpadata.Lpa{
				Voucher: lpadata.Voucher{FirstNames: "V", LastName: "W"},
			},
			FirstNames: "A",
			LastName:   "B",
		}).
		Return(nil)

	err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{
		LpaID:      "lpa-id",
		FirstNames: "A",
		LastName:   "B",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourNameWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := ConfirmYourName(nil, lpaStoreResolvingService, nil)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourNameWhenTemplateErrors(t *testing.T) {
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

	err := ConfirmYourName(template.Execute, lpaStoreResolvingService, nil)(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID: "lpa-id",
			Tasks: voucherdata.Tasks{ConfirmYourName: task.StateInProgress},
		}).
		Return(nil)

	err := ConfirmYourName(nil, nil, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourNameWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourName(nil, nil, voucherStore)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
