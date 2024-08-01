package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAreYouApplyingForFeeDiscountOrExemption(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &areYouApplyingForFeeDiscountOrExemptionData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouApplyingForFeeDiscountOrExemptionWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &areYouApplyingForFeeDiscountOrExemptionData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(expectedError)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouApplyingForFeeDiscountOrExemption(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.DonorProvidedDetails{LpaID: "lpa-id", Donor: donordata.Donor{Email: "a@b.com"}}

	payer := newMockHandler(t)
	payer.EXPECT().
		Execute(testAppData, w, r, donor).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: donordata.Donor{Email: "a@b.com"},
			Tasks: donordata.DonorTasks{PayForLpa: task.PaymentStateInProgress},
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, payer.Execute, donorStore)(testAppData, w, r, donor)
	assert.Nil(t, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenDonorStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenPayerErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	payer := newMockHandler(t)
	payer.EXPECT().
		Execute(testAppData, w, r, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, payer.Execute, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenYes(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID: "lpa-id",
			Donor: donordata.Donor{Email: "a@b.com"},
			Tasks: donordata.DonorTasks{PayForLpa: task.PaymentStateInProgress},
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id", Donor: donordata.Donor{Email: "a@b.com"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WhichFeeTypeAreYouApplyingFor.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "whetherApplyingForDifferentFeeType"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *areYouApplyingForFeeDiscountOrExemptionData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{LpaID: "lpa-id", Donor: donordata.Donor{Email: "a@b.com"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
