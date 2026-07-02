package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCompletingYourIdentityConfirmation(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &completingYourIdentityConfirmationData{
			App:  testAppData,
			Form: forms.NewEnumForm[howYouWillConfirmYourIdentity]("howYouWouldLikeToContinue", howYouWillConfirmYourIdentityValues),
		}).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCompletingYourIdentityConfirmationWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	assert.Equal(t, expectedError, err)
}

func TestPostCompletingYourIdentityConfirmation(t *testing.T) {
	testCases := map[string]struct {
		how      howYouWillConfirmYourIdentity
		provided *certificateproviderdata.Provided
		redirect certificateprovider.Path
	}{
		"post office successful": {
			how:      howYouWillConfirmYourIdentityPostOfficeSuccessfully,
			provided: &certificateproviderdata.Provided{LpaID: "lpa-id"},
			redirect: certificateprovider.PathIdentityWithOneLogin,
		},
		"one login": {
			how:      howYouWillConfirmYourIdentityOneLogin,
			provided: &certificateproviderdata.Provided{LpaID: "lpa-id"},
			redirect: certificateprovider.PathIdentityWithOneLogin,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"enum": {tc.how.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := CompletingYourIdentityConfirmation(nil)(testAppData, w, r, tc.provided, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostCompletingYourIdentityConfirmationWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *completingYourIdentityConfirmationData) bool {
			return assert.Equal(t, []forms.Field{data.Form.Enum.Field}, data.Form.Errors) &&
				assert.Equal(t, "errorSelect:Label=howYouWouldLikeToContinue", data.Form.Enum.Error.Format(testAppData.Localizer))
		})).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
