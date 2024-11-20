package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostHowWillYouConfirmYourIdentity(t *testing.T) {
	testCases := map[string]struct {
		how      howYouWillConfirmYourIdentity
		provided *donordata.Provided
		redirect donor.Path
	}{
		"post office successful": {
			how:      howYouWillConfirmYourIdentityPostOfficeSuccessfully,
			provided: &donordata.Provided{LpaID: "lpa-id"},
			redirect: donor.PathIdentityWithOneLogin,
		},
		"one login": {
			how:      howYouWillConfirmYourIdentityOneLogin,
			provided: &donordata.Provided{LpaID: "lpa-id"},
			redirect: donor.PathIdentityWithOneLogin,
		},
		"delete": {
			how:      howYouWillConfirmYourIdentityWithdraw,
			provided: &donordata.Provided{LpaID: "lpa-id"},
			redirect: donor.PathDeleteThisLpa,
		},
		"withdraw": {
			how:      howYouWillConfirmYourIdentityWithdraw,
			provided: &donordata.Provided{LpaID: "lpa-id", WitnessedByCertificateProviderAt: time.Now()},
			redirect: donor.PathWithdrawThisLpa,
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

			err := HowWillYouConfirmYourIdentity(nil, nil)(testAppData, w, r, tc.provided)
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

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Tasks: donordata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
		}).
		Return(nil)

	err := HowWillYouConfirmYourIdentity(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostHowWillYouConfirmYourIdentityWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {howYouWillConfirmYourIdentityAtPostOffice.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := HowWillYouConfirmYourIdentity(nil, donorStore)(testAppData, w, r, &donordata.Provided{})
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

	err := HowWillYouConfirmYourIdentity(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
