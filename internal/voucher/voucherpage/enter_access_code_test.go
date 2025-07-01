package voucherpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnterAccessCode(t *testing.T) {
	accessCode := accesscodedata.Link{LpaKey: dynamo.LpaKey("hi")}
	session := &sesh.LoginSession{Email: "a@example.com"}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Create(r.Context(), accessCode, "a@example.com").
		Return(nil, nil)

	err := EnterAccessCode(voucherStore)(testAppData, w, r, session, nil, accessCode)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestEnterAccessCodeOnVoucherStoreError(t *testing.T) {
	accessCode := accesscodedata.Link{}
	session := &sesh.LoginSession{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterAccessCode(voucherStore)(testAppData, w, r, session, nil, accessCode)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
