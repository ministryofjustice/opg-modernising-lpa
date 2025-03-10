package donorpage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetResendVoucherAccessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &resendVoucherAccessCodeData{
			App: testAppData,
		}).
		Return(nil)

	err := ResendVoucherAccessCode(template.Execute, &mockShareCodeSender{})(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetResendVoucherAccessCodeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ResendVoucherAccessCode(template.Execute, &mockShareCodeSender{})(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostResendVoucherAccessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	provided := &donordata.Provided{
		LpaID:            "lpa-id",
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
	}

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendVoucherAccessCode(r.Context(), provided, testAppData).
		Return(nil)

	err := ResendVoucherAccessCode(nil, shareCodeSender)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWeHaveContactedVoucher.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostResendVoucherAccessCodeWhenSendErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{Donor: donordata.Donor{FirstNames: "john"}}

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendVoucherAccessCode(r.Context(), donor, testAppData).
		Return(expectedError)

	err := ResendVoucherAccessCode(nil, shareCodeSender)(testAppData, w, r, donor)

	assert.Equal(t, expectedError, err)
}
