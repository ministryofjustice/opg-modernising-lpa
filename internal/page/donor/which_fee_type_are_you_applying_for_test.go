package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhichFeeTypeAreYouApplyingFor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{},
			Options: page.FeeTypeValues,
		}).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhichFeeTypeAreYouApplyingForWithLpaData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{FeeType: page.HalfFee},
			Options: page.FeeTypeValues,
		}).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &page.Lpa{FeeType: page.HalfFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhichFeeTypeAreYouApplyingForOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{},
			Options: page.FeeTypeValues,
		}).
		Return(expectedError)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhichFeeTypeAreYouApplyingFor(t *testing.T) {
	testcases := map[page.FeeType]page.LpaPath{
		page.HalfFee:              page.Paths.EvidenceRequired,
		page.NoFee:                page.Paths.EvidenceRequired,
		page.HardshipFee:          page.Paths.EvidenceRequired,
		page.RepeatApplicationFee: page.Paths.PreviousApplicationNumber,
	}

	for feeType, redirect := range testcases {
		t.Run(feeType.String(), func(t *testing.T) {
			form := url.Values{
				"fee-type": {feeType.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), &page.Lpa{ID: "lpa-id", FeeType: feeType}).
				Return(nil)

			err := WhichFeeTypeAreYouApplyingFor(nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostWhichFeeTypeAreYouApplyingForOnStoreError(t *testing.T) {
	form := url.Values{
		"fee-type": {page.HalfFee.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{ID: "lpa-id", FeeType: page.HalfFee}).
		Return(expectedError)

	err := WhichFeeTypeAreYouApplyingFor(nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhichFeeTypeAreYouApplyingForOnInvalidForm(t *testing.T) {
	form := url.Values{
		"fee-type": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("fee-type", validation.SelectError{Label: "whichFeeTypeYouAreApplyingFor"})

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *whichFeeTypeAreYouApplyingForData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
