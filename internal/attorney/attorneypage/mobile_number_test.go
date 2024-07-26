package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMobileNumber(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
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
				Execute(w, &mobileNumberData{
					App:  tc.appData,
					Form: &mobileNumberForm{},
				}).
				Return(nil)

			err := MobileNumber(template.Execute, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetMobileNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData  page.AppData
		attorney *actor.AttorneyProvidedDetails
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
				Execute(w, &mobileNumberData{
					App: tc.appData,
					Form: &mobileNumberForm{
						Mobile: "07535111222",
					},
				}).
				Return(nil)

			err := MobileNumber(template.Execute, nil)(tc.appData, w, r, &actor.AttorneyProvidedDetails{Mobile: "07535111222"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetMobileNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := MobileNumber(template.Execute, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostMobileNumber(t *testing.T) {
	testCases := map[string]struct {
		form            url.Values
		attorney        *actor.AttorneyProvidedDetails
		updatedAttorney *actor.AttorneyProvidedDetails
		appData         page.AppData
	}{
		"attorney": {
			form: url.Values{
				"mobile": {"07535111222"},
			},
			attorney: &actor.AttorneyProvidedDetails{LpaID: "lpa-id"},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID:  "lpa-id",
				Mobile: "07535111222",
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskInProgress,
				},
			},
			appData: testAppData,
		},
		"attorney empty": {
			appData: testAppData,
			attorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskCompleted,
				},
			},
		},
		"replacement attorney": {
			form: url.Values{
				"mobile": {"07535111222"},
			},
			attorney: &actor.AttorneyProvidedDetails{LpaID: "lpa-id"},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID:  "lpa-id",
				Mobile: "07535111222",
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskInProgress,
				},
			},
			appData: testReplacementAppData,
		},
		"replacement attorney empty": {
			appData:  testReplacementAppData,
			attorney: &actor.AttorneyProvidedDetails{LpaID: "lpa-id"},
			updatedAttorney: &actor.AttorneyProvidedDetails{
				LpaID: "lpa-id",
				Tasks: actor.AttorneyTasks{
					ConfirmYourDetails: actor.TaskInProgress,
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

			err := MobileNumber(nil, attorneyStore)(tc.appData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.YourPreferredLanguage.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostMobileNumberWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"mobile": {"0123456"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *mobileNumberData) bool {
		return assert.Equal(t, validation.With("mobile", validation.PhoneError{Tmpl: "errorMobile", Label: "mobile"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *mobileNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := MobileNumber(template.Execute, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostMobileNumberWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := MobileNumber(nil, attorneyStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadMobileNumberForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"mobile": {"07535111222"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readMobileNumberForm(r)

	assert.Equal("07535111222", result.Mobile)
}

func TestMobileNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *mobileNumberForm
		errors validation.List
	}{
		"valid": {
			form: &mobileNumberForm{
				Mobile: "07535999222",
			},
		},
		"missing": {
			form: &mobileNumberForm{},
		},
		"invalid-mobile-format": {
			form: &mobileNumberForm{
				Mobile: "123",
			},
			errors: validation.With("mobile", validation.PhoneError{Tmpl: "errorMobile", Label: "mobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
