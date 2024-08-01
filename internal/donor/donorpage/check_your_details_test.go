package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

func TestGetCheckYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.DonorProvidedDetails{}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourDetailsData{App: testAppData, Donor: donor}).
		Return(nil)

	err := CheckYourDetails(template.Execute)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &checkYourDetailsData{App: testAppData, Donor: &donordata.DonorProvidedDetails{}}).
		Return(expectedError)

	err := CheckYourDetails(template.Execute)(testAppData, w, r, &donordata.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostCheckYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := CheckYourDetails(nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID: "lpa-id",
		Tasks: donordata.DonorTasks{
			PayForLpa: task.PaymentStateCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WeHaveContactedVoucher.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCheckYourDetailsWhenUnpaid(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := CheckYourDetails(nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WeHaveReceivedVoucherDetails.Format("lpa-id"), resp.Header.Get("Location"))
}
