package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
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
			Form: form.NewEmptySelectForm[howYouWillConfirmYourIdentity](howYouWillConfirmYourIdentityValues, "howYouWouldLikeToContinue"),
		}).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{})
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

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{})
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
				form.FieldNames.Select: {tc.how.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := CompletingYourIdentityConfirmation(nil)(testAppData, w, r, tc.provided)
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
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "howYouWouldLikeToContinue"}), data.Errors)
		})).
		Return(nil)

	err := CompletingYourIdentityConfirmation(template.Execute)(testAppData, w, r, &certificateproviderdata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
