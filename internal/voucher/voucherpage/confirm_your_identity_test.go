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

func TestGetConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaID: "lpa-id"}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmYourIdentityData{
			App: testAppData,
			Lpa: &lpadata.Lpa{LpaID: "lpa-id"},
		}).
		Return(nil)

	err := ConfirmYourIdentity(template.Execute, nil, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourIdentityWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(nil, expectedError)

	err := ConfirmYourIdentity(nil, nil, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestGetConfirmYourIdentityWhenTemplateErrors(t *testing.T) {
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

	err := ConfirmYourIdentity(template.Execute, nil, lpaStoreResolvingService)(testAppData, w, r, &voucherdata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), &voucherdata.Provided{
			LpaID: "lpa-id",
			Tasks: voucherdata.Tasks{ConfirmYourIdentity: task.IdentityStateInProgress},
		}).
		Return(nil)

	err := ConfirmYourIdentity(nil, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathIdentityWithOneLogin.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourIdentityWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourIdentity(nil, voucherStore, nil)(testAppData, w, r, &voucherdata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}
