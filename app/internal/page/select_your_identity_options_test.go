package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSelectYourIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:  appData,
			Form: &selectYourIdentityOptionsForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SelectYourIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SelectYourIdentityOptions(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetSelectYourIdentityOptionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			IdentityOptions: IdentityOptions{
				Selected: []IdentityOption{Passport},
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:  appData,
			Form: &selectYourIdentityOptionsForm{Options: []IdentityOption{Passport}},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SelectYourIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetSelectYourIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SelectYourIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostSelectYourIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			IdentityOptions: IdentityOptions{
				Selected: []IdentityOption{Passport, DwpAccount, UtilityBill},
				First:    Passport,
				Second:   DwpAccount,
			},
			Tasks: Tasks{
				ConfirmYourIdentityAndSign: TaskInProgress,
			},
		}).
		Return(nil)

	form := url.Values{
		"options": {"passport", "dwp account", "utility bill"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := SelectYourIdentityOptions(nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, yourChosenIdentityOptionsPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"options": {"passport", "dwp account", "utility bill"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := SelectYourIdentityOptions(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostSelectYourIdentityOptionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &selectYourIdentityOptionsData{
			App:  appData,
			Form: &selectYourIdentityOptionsForm{Options: []IdentityOption{}},
			Errors: map[string]string{
				"options": "selectAtLeastThreeIdentityOptions",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := SelectYourIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadSelectYourIdentityOptionsForm(t *testing.T) {
	form := url.Values{
		"options": {"passport", "driving licence", "council tax bill"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readSelectYourIdentityOptionsForm(r)

	assert.Equal(t, []IdentityOption{Passport, DrivingLicence, CouncilTaxBill}, result.Options)
}

func TestSelectYourIdentityOptionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *selectYourIdentityOptionsForm
		errors map[string]string
	}{
		"all": {
			form: &selectYourIdentityOptionsForm{
				Options: []IdentityOption{Yoti, Passport, DrivingLicence, GovernmentGatewayAccount, DwpAccount, OnlineBankAccount, UtilityBill, CouncilTaxBill},
				First:   Yoti,
				Second:  DwpAccount,
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &selectYourIdentityOptionsForm{},
			errors: map[string]string{
				"options": "selectAtLeastThreeIdentityOptions",
			},
		},
		"too-few": {
			form: &selectYourIdentityOptionsForm{
				Options: []IdentityOption{Passport, DwpAccount},
			},
			errors: map[string]string{
				"options": "selectAtLeastThreeIdentityOptions",
			},
		},
		"incompatible-second": {
			form: &selectYourIdentityOptionsForm{
				Options: []IdentityOption{Yoti, Passport, DrivingLicence},
				First:   Yoti,
				Second:  IdentityOptionUnknown,
			},
			errors: map[string]string{
				"options": "selectMoreOptions",
			},
		},
		"invalid": {
			form: &selectYourIdentityOptionsForm{
				Options: []IdentityOption{Passport, IdentityOptionUnknown},
			},
			errors: map[string]string{
				"options": "selectValidIdentityOption",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
