package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPreviousFee(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &previousFeeData{
			App:     testAppData,
			Form:    &previousFeeForm{},
			Options: pay.PreviousFeeValues,
		}).
		Return(nil)

	err := PreviousFee(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousFeeFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &previousFeeData{
			App: testAppData,
			Form: &previousFeeForm{
				PreviousFee: pay.PreviousFeeHalf,
			},
			Options: pay.PreviousFeeValues,
		}).
		Return(nil)

	err := PreviousFee(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PreviousFee: pay.PreviousFeeHalf})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetPreviousFeeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := PreviousFee(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPreviousFeeWhenFullFee(t *testing.T) {
	form := url.Values{
		"previous-fee": {pay.PreviousFeeFull.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donor := &donordata.Provided{
		LpaID:       "lpa-id",
		PreviousFee: pay.PreviousFeeFull,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), donor).
		Return(nil)

	payer := newMockHandler(t)
	payer.EXPECT().
		Execute(testAppData, w, r, donor).
		Return(nil)

	err := PreviousFee(nil, payer.Execute, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	assert.Nil(t, err)
}

func TestPostPreviousFeeWhenOtherFee(t *testing.T) {
	form := url.Values{
		"previous-fee": {pay.PreviousFeeHalf.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:       "lpa-id",
			PreviousFee: pay.PreviousFeeHalf,
		}).
		Return(nil)

	err := PreviousFee(nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEvidenceRequired.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostPreviousFeeWhenNotChanged(t *testing.T) {
	form := url.Values{
		"previous-fee": {pay.PreviousFeeHalf.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := PreviousFee(nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:       "lpa-id",
		PreviousFee: pay.PreviousFeeHalf,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathEvidenceRequired.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostPreviousFeeWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"previous-fee": {pay.PreviousFeeHalf.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := PreviousFee(nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostPreviousFeeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *previousFeeData) bool {
			return assert.Equal(t, validation.With("previous-fee", validation.SelectError{Label: "howMuchYouPreviouslyPaid"}), data.Errors)
		})).
		Return(nil)

	err := PreviousFee(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadPreviousFeeForm(t *testing.T) {
	form := url.Values{
		"previous-fee": {pay.PreviousFeeHalf.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readPreviousFeeForm(r)
	assert.Equal(t, pay.PreviousFeeHalf, result.PreviousFee)
}

func TestPreviousFeeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *previousFeeForm
		errors validation.List
	}{
		"valid": {
			form: &previousFeeForm{
				PreviousFee: pay.PreviousFeeFull,
			},
		},
		"invalid": {
			form:   &previousFeeForm{},
			errors: validation.With("previous-fee", validation.SelectError{Label: "howMuchYouPreviouslyPaid"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
