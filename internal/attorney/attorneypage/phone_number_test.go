package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPhoneNumber(t *testing.T) {
	testcases := map[string]struct {
		appData appcontext.Data
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &phoneNumberData{
					App:  tc.appData,
					Form: &phoneNumberForm{},
				}).
				Return(nil)

			err := PhoneNumber(template.Execute, nil)(tc.appData, w, r, &attorneydata.Provided{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPhoneNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData  appcontext.Data
		attorney *attorneydata.Provided
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &phoneNumberData{
					App: tc.appData,
					Form: &phoneNumberForm{
						Phone: "07535111222",
					},
				}).
				Return(nil)

			err := PhoneNumber(template.Execute, nil)(tc.appData, w, r, &attorneydata.Provided{Telephone: "07535111222"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPhoneNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := PhoneNumber(template.Execute, nil)(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPhoneNumber(t *testing.T) {
	testCases := map[string]struct {
		form            url.Values
		attorney        *attorneydata.Provided
		updatedAttorney *attorneydata.Provided
		appData         appcontext.Data
	}{
		"attorney": {
			form: url.Values{
				"phone": {"07535111222"},
			},
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID:     "lpa-id",
				Telephone: "07535111222",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
			appData: testAppData,
		},
		"attorney empty": {
			appData: testAppData,
			attorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
		},
		"replacement attorney": {
			form: url.Values{
				"phone": {"07535111222"},
			},
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID:     "lpa-id",
				Telephone: "07535111222",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
			appData: testReplacementAppData,
		},
		"replacement attorney empty": {
			appData:  testReplacementAppData,
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), tc.updatedAttorney).
				Return(nil)

			err := PhoneNumber(nil, attorneyStore)(tc.appData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathYourPreferredLanguage.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostPhoneNumberWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"phone": {"abcd"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *phoneNumberData) bool {
		return assert.Equal(t, validation.With("phone", validation.PhoneError{Tmpl: "errorTelephone", Label: "phone"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *phoneNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := PhoneNumber(template.Execute, nil)(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPhoneNumberWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		"phone": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := PhoneNumber(nil, attorneyStore)(testAppData, w, r, &attorneydata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadPhoneNumberForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"phone": {"07535111222"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readPhoneNumberForm(r)

	assert.Equal("07535111222", result.Phone)
}

func TestPhoneNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *phoneNumberForm
		errors validation.List
	}{
		"valid": {
			form: &phoneNumberForm{
				Phone: "07535999222",
			},
		},
		"missing": {
			form: &phoneNumberForm{},
		},
		"invalid-phone-format": {
			form: &phoneNumberForm{
				Phone: "123",
			},
			errors: validation.With("phone", validation.PhoneError{Tmpl: "errorTelephone", Label: "phone"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
