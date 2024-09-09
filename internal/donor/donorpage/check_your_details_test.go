package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

func TestGetCheckYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourDetailsData{App: testAppData, Donor: donor}).
		Return(nil)

	err := CheckYourDetails(template.Execute, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourDetailsData{App: testAppData, Donor: &donordata.Provided{}}).
		Return(expectedError)

	err := CheckYourDetails(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostCheckYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{
			PayForLpa: task.PaymentStateCompleted,
		},
	}

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendVoucherAccessCode(r.Context(), provided, testAppData).
		Return(nil)

	err := CheckYourDetails(nil, shareCodeSender)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWeHaveContactedVoucher.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCheckYourDetailsWhenUnpaid(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := CheckYourDetails(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWeHaveReceivedVoucherDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCheckYourDetailsWhenShareCodeStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{
			PayForLpa: task.PaymentStateCompleted,
		},
	}

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendVoucherAccessCode(r.Context(), provided, testAppData).
		Return(expectedError)

	err := CheckYourDetails(nil, shareCodeSender)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
