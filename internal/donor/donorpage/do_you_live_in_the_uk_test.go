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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDoYouLiveInTheUK(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &doYouLiveInTheUKData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := DoYouLiveInTheUK(template.Execute)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDoYouLiveInTheUKWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := DoYouLiveInTheUK(template.Execute)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostDoYouLiveInTheUK(t *testing.T) {
	testCases := []struct {
		yesNo    form.YesNo
		redirect donor.Path
	}{
		{
			yesNo:    form.Yes,
			redirect: donor.PathYourAddress,
		},
		{
			yesNo:    form.No,
			redirect: donor.PathWhatCountryDoYouLiveIn,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.yesNo.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := DoYouLiveInTheUK(nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostDoYouLiveInTheUKWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *doYouLiveInTheUKData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesIfYouLiveInTheUK"}), data.Errors)
		})).
		Return(nil)

	err := DoYouLiveInTheUK(template.Execute)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
