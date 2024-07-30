package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	pay "github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhichFeeTypeAreYouApplyingFor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{},
			Options: pay.FeeTypeValues,
		}).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhichFeeTypeAreYouApplyingForWithLpaData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{FeeType: pay.HalfFee},
			Options: pay.FeeTypeValues,
		}).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{FeeType: pay.HalfFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhichFeeTypeAreYouApplyingForOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whichFeeTypeAreYouApplyingForData{
			App:     testAppData,
			Form:    &whichFeeTypeAreYouApplyingForForm{},
			Options: pay.FeeTypeValues,
		}).
		Return(expectedError)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhichFeeTypeAreYouApplyingFor(t *testing.T) {
	testcases := map[pay.FeeType]page.LpaPath{
		pay.HalfFee:              page.Paths.EvidenceRequired,
		pay.NoFee:                page.Paths.EvidenceRequired,
		pay.HardshipFee:          page.Paths.EvidenceRequired,
		pay.RepeatApplicationFee: page.Paths.PreviousApplicationNumber,
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
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", FeeType: feeType}).
				Return(nil)

			err := WhichFeeTypeAreYouApplyingFor(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, redirect.Format("lpa-id"), resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostWhichFeeTypeAreYouApplyingForOnStoreError(t *testing.T) {
	form := url.Values{
		"fee-type": {pay.HalfFee.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", FeeType: pay.HalfFee}).
		Return(expectedError)

	err := WhichFeeTypeAreYouApplyingFor(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
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
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whichFeeTypeAreYouApplyingForData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := WhichFeeTypeAreYouApplyingFor(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
