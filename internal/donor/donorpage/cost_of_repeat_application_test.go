package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCostOfRepeatApplication(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &costOfRepeatApplicationData{
			App:  testAppData,
			Form: form.NewEmptySelectForm[pay.CostOfRepeatApplication](pay.CostOfRepeatApplicationValues, "whichFeeYouAreEligibleToPay"),
		}).
		Return(nil)

	err := CostOfRepeatApplication(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCostOfRepeatApplicationFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &costOfRepeatApplicationData{
			App:  testAppData,
			Form: form.NewSelectForm(pay.CostOfRepeatApplicationHalfFee, pay.CostOfRepeatApplicationValues, "whichFeeYouAreEligibleToPay"),
		}).
		Return(nil)

	err := CostOfRepeatApplication(template.Execute, nil)(testAppData, w, r, &donordata.Provided{CostOfRepeatApplication: pay.CostOfRepeatApplicationHalfFee})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCostOfRepeatApplicationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := CostOfRepeatApplication(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCostOfRepeatApplication(t *testing.T) {
	for cost, path := range map[pay.CostOfRepeatApplication]donor.Path{
		pay.CostOfRepeatApplicationNoFee:   donor.PathWhatHappensNextRepeatApplicationNoFee,
		pay.CostOfRepeatApplicationHalfFee: donor.PathPreviousFee,
	} {
		t.Run(cost.String(), func(t *testing.T) {
			form := url.Values{
				form.FieldNames.Select: {cost.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			provided := &donordata.Provided{
				LpaID:                   "lpa-id",
				CostOfRepeatApplication: cost,
				Tasks:                   donordata.Tasks{PayForLpa: task.PaymentStatePending},
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), provided).
				Return(nil)

			err := CostOfRepeatApplication(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, path.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCostOfRepeatApplicationWhenNotChanged(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {pay.CostOfRepeatApplicationHalfFee.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := CostOfRepeatApplication(nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:                   "lpa-id",
		CostOfRepeatApplication: pay.CostOfRepeatApplicationHalfFee,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathPreviousFee.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCostOfRepeatApplicationWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {pay.CostOfRepeatApplicationHalfFee.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := CostOfRepeatApplication(nil, donorStore)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostCostOfRepeatApplicationWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *costOfRepeatApplicationData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "whichFeeYouAreEligibleToPay"}), data.Errors)
		})).
		Return(nil)

	err := CostOfRepeatApplication(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
