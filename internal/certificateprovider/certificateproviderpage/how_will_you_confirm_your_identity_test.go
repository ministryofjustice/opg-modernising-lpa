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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowWillYouConfirmYourIdentity(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &howWillYouConfirmYourIdentityData{
			App:  testAppData,
			Form: form.NewEmptySelectForm[howYouWillConfirmYourIdentity](howYouWillConfirmYourIdentityValues, "howYouWillConfirmYourIdentity"),
		}).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetHowWillYouConfirmYourIdentityWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	assert.Equal(t, expectedError, err)
}

func TestPostHowWillYouConfirmYourIdentity(t *testing.T) {
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

			err := HowWillYouConfirmYourIdentity(nil, nil)(testAppData, w, r, tc.provided, nil)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostHowWillYouConfirmYourIdentityWhenAtPostOfficeSelected(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &certificateproviderdata.Provided{
			LpaID: "lpa-id",
			Tasks: certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
		}).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{LpaID: "lpa-id"}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWillYouConfirmYourIdentityWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWillYouConfirmYourIdentity(nil, certificateProviderStore)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	assert.ErrorIs(t, err, expectedError)
}

func TestPostHowWillYouConfirmYourIdentityWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *howWillYouConfirmYourIdentityData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "howYouWillConfirmYourIdentity"}), data.Errors)
		})).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
