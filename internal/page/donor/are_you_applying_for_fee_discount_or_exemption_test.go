package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAreYouApplyingForFeeDiscountOrExemption(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouApplyingForFeeDiscountOrExemption{
			App:     testAppData,
			Options: form.YesNoValues,
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAreYouApplyingForFeeDiscountOrExemptionWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &areYouApplyingForFeeDiscountOrExemption{
			App:     testAppData,
			Options: form.YesNoValues,
		}).
		Return(expectedError)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAreYouApplyingForFeeDiscountOrExemption(t *testing.T) {
	f := url.Values{
		"yes-no": {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}}

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, lpa).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:    "lpa-id",
			Donor: actor.Donor{Email: "a@b.com"},
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskInProgress},
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, payer, donorStore)(testAppData, w, r, lpa)
	assert.Nil(t, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenDonorStoreErrors(t *testing.T) {
	form := url.Values{
		"yes-no": {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, nil, donorStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenPayerErrors(t *testing.T) {
	form := url.Values{
		"yes-no": {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	payer := newMockPayer(t)
	payer.
		On("Pay", testAppData, w, r, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, payer, donorStore)(testAppData, w, r, &page.Lpa{})
	assert.Equal(t, expectedError, err)
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenYes(t *testing.T) {
	f := url.Values{
		"yes-no": {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:    "lpa-id",
			Donor: actor.Donor{Email: "a@b.com"},
			Tasks: page.Tasks{PayForLpa: actor.PaymentTaskInProgress},
		}).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(nil, nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.WhichFeeTypeAreYouApplyingFor.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostAreYouApplyingForFeeDiscountOrExemptionWhenValidationError(t *testing.T) {
	form := url.Values{
		"yes-no": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("yes-no", validation.SelectError{Label: "whetherApplyingForDifferentFeeType"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *areYouApplyingForFeeDiscountOrExemption) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := AreYouApplyingForFeeDiscountOrExemption(template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id", Donor: actor.Donor{Email: "a@b.com"}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
