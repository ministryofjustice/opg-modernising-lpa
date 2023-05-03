package attorney

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
		lpa     *page.Lpa
		appData page.AppData
	}{
		"attorney": {
			lpa: &page.Lpa{
				ID:                      "lpa-id",
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			appData: testAppData,
		},
		"replacement attorney": {
			lpa: &page.Lpa{
				ID:                                 "lpa-id",
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &mobileNumberData{
					App:  tc.appData,
					Lpa:  tc.lpa,
					Form: &mobileNumberForm{},
				}).
				Return(nil)

			err := MobileNumber(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetMobileNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Mobile: "07535111222",
					},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Mobile: "07535111222",
					},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &mobileNumberData{
					App: tc.appData,
					Lpa: tc.lpa,
					Form: &mobileNumberForm{
						Mobile: "07535111222",
					},
				}).
				Return(nil)

			err := MobileNumber(template.Execute, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetMobileNumberWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := MobileNumber(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetMobileNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		ID: "lpa-id",
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &mobileNumberData{
			App:  testAppData,
			Lpa:  lpa,
			Form: &mobileNumberForm{},
		}).
		Return(expectedError)

	err := MobileNumber(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostMobileNumber(t *testing.T) {
	testCases := map[string]struct {
		form       url.Values
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		appData    page.AppData
	}{
		"attorney": {
			form: url.Values{
				"mobile": {"07535111222"},
			},
			lpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Mobile: "07535111222",
					},
				},
				AttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testAppData,
		},
		"attorney empty": {
			lpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
				AttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testAppData,
		},
		"replacement attorney": {
			form: url.Values{
				"mobile": {"07535111222"},
			},
			lpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Mobile: "07535111222",
					},
				},
				ReplacementAttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testReplacementAppData,
		},
		"replacement attorney empty": {
			lpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
				ReplacementAttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)
			lpaStore.
				On("Put", r.Context(), tc.updatedLpa).
				Return(nil)

			err := MobileNumber(nil, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.ReadTheLpa, resp.Header.Get("Location"))
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

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	dataMatcher := func(t *testing.T, data *mobileNumberData) bool {
		return assert.Equal(t, validation.With("mobile", validation.MobileError{Label: "mobile"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *mobileNumberData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := MobileNumber(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostMobileNumberWhenLpaStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ID: "lpa-id"}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := MobileNumber(nil, lpaStore)(testAppData, w, r)
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
			errors: validation.With("mobile", validation.MobileError{Label: "mobile"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
