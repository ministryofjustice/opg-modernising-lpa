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

func TestGetChooseAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:  appData,
			Form: &chooseAttorneysAddressForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestGetChooseAttorneysAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetChooseAttorneysAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	address := Address{Line1: "abc"}
	dataStore := &mockDataStore{
		data: Lpa{
			Attorney: Attorney{
				Address: address,
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Attorney: Attorney{
				Address: address,
			},
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestGetChooseAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &Address{},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChooseAttorneysAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:  appData,
			Form: &chooseAttorneysAddressForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Attorney: Attorney{
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

	err := ChooseAttorneysAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostChooseAttorneysAddressManualWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Attorney: Attorney{
				Address: Address{
					Line1:      "a",
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
		}).
		Return(expectedError)

	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseAttorneysAddress(nil, nil, nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostChooseAttorneysAddressManualFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			Attorney: Attorney{
				FirstNames: "John",
				Address:    Address{Line1: "abc"},
			},
			WhoFor: "me",
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Attorney: Attorney{
				FirstNames: "John",
				Address: Address{
					Line1:      "a",
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
			WhoFor: "me",
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

	err := ChooseAttorneysAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostChooseAttorneysAddressManualWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestPostChooseAttorneysAddressSelect(t *testing.T) {
	w := httptest.NewRecorder()

	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Attorney: Attorney{
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

	err := ChooseAttorneysAddress(nil, nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostChooseAttorneysAddressSelectWhenValidationError(t *testing.T) {
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
		On("Get", mock.Anything, "session-id").
		Return(nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseAttorneysAddress(nil, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseAttorneysAddressLookup(t *testing.T) {
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
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseAttorneysAddress(nil, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChooseAttorneysAddressLookupError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return([]Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseAttorneysAddress(logger, template.Func, addressClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChooseAttorneysAddressLookupWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action": {"lookup"},
	}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
				Action: "lookup",
			},
			Errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseAttorneysAddress(nil, template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadChooseAttorneysAddressForm(t *testing.T) {
	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	testCases := map[string]struct {
		form   url.Values
		result *chooseAttorneysAddressForm
	}{
		"lookup": {
			form: url.Values{
				"action":          {"lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &chooseAttorneysAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"select": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &chooseAttorneysAddressForm{
				Action:  "select",
				Address: expectedAddress,
			},
		},
		"select-not-selected": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {""},
			},
			result: &chooseAttorneysAddressForm{
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
			result: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			actual := readChooseAttorneysAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestChooseAttorneysAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *chooseAttorneysAddressForm
		errors map[string]string
	}{
		"lookup-valid": {
			form: &chooseAttorneysAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			errors: map[string]string{},
		},
		"lookup-missing-postcode": {
			form: &chooseAttorneysAddressForm{
				Action: "lookup",
			},
			errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		},
		"select-valid": {
			form: &chooseAttorneysAddressForm{
				Action:  "select",
				Address: &Address{},
			},
			errors: map[string]string{},
		},
		"select-not-selected": {
			form: &chooseAttorneysAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: map[string]string{
				"select-address": "selectAddress",
			},
		},
		"manual-valid": {
			form: &chooseAttorneysAddressForm{
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
			form: &chooseAttorneysAddressForm{
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
