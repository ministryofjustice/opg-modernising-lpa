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

type mockAddressClient struct {
	mock.Mock
}

func (m *mockAddressClient) LookupPostcode(postcode string) ([]Address, error) {
	args := m.Called(postcode)
	return args.Get(0).([]Address), args.Error(1)
}

func TestGetDonorAddress(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App:  appData,
			Form: &donorAddressForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	DonorAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestGetDonorAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action:  "manual",
				Address: &Address{},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	DonorAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetDonorAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App:  appData,
			Form: &donorAddressForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	DonorAddress(logger, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}

func TestPostDonorAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Donor: Donor{
				Address: Address{
					Line1:      "a",
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
		}).
		Return(nil)

	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, howWouldYouLikeToBeContactedPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostDonorAddressManualWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action: "manual",
				Address: &Address{
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
			Errors: map[string]string{
				"address-line-1": "enterAddress",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestPostDonorAddressSelect(t *testing.T) {
	w := httptest.NewRecorder()

	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Donor: Donor{
				Address: *expectedAddress,
			},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {expectedAddress.Encode()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, howWouldYouLikeToBeContactedPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostDonorAddressSelectWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	addresses := []Address{
		{Line1: "a", Line2: "b"},
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors: map[string]string{
				"select-address": "selectAddress",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostDonorAddressLookup(t *testing.T) {
	w := httptest.NewRecorder()

	addresses := []Address{
		{Line1: "a", Line2: "b"},
	}

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(addresses, nil)

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    map[string]string{},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostDonorAddressLookupError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return([]Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []Address{},
			Errors: map[string]string{
				"lookup-postcode": "couldNotLookupPostcode",
			},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(logger, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostDonorAddressLookupWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action": {"lookup"},
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &donorAddressData{
			App: appData,
			Form: &donorAddressForm{
				Action: "lookup",
			},
			Errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	DonorAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadDonorAddressForm(t *testing.T) {
	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	testCases := map[string]struct {
		form   url.Values
		result *donorAddressForm
	}{
		"lookup": {
			form: url.Values{
				"action":          {"lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &donorAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"select": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &donorAddressForm{
				Action:  "select",
				Address: expectedAddress,
			},
		},
		"select-not-selected": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {""},
			},
			result: &donorAddressForm{
				Action:  "select",
				Address: nil,
			},
		},
		"manual": {
			form: url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-line-2":   {"b"},
				"address-town":     {"c"},
				"address-postcode": {"d"},
			},
			result: &donorAddressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			actual := readDonorAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestDonorAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *donorAddressForm
		errors map[string]string
	}{
		"lookup-valid": {
			form: &donorAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			errors: map[string]string{},
		},
		"lookup-missing-postcode": {
			form: &donorAddressForm{
				Action: "lookup",
			},
			errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		},
		"select-valid": {
			form: &donorAddressForm{
				Action:  "select",
				Address: &Address{},
			},
			errors: map[string]string{},
		},
		"select-not-selected": {
			form: &donorAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: map[string]string{
				"select-address": "selectAddress",
			},
		},
		"manual-valid": {
			form: &donorAddressForm{
				Action: "manual",
				Address: &Address{
					Line1:      "a",
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
			errors: map[string]string{},
		},
		"manual-missing-all": {
			form: &donorAddressForm{
				Action:  "manual",
				Address: &Address{},
			},
			errors: map[string]string{
				"address-line-1":   "enterAddress",
				"address-town":     "enterTownOrCity",
				"address-postcode": "enterPostcode",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
