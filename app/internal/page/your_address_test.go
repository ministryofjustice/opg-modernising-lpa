package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/ordnance_survey"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAddressClient struct {
	mock.Mock
}

func (m *mockAddressClient) LookupPostcode(postcode string) (ordnance_survey.PostcodeLookupResponse, error) {
	args := m.Called(postcode)
	return args.Get(0).(ordnance_survey.PostcodeLookupResponse), args.Error(1)
}

func TestGetYourAddress(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App:  appData,
			Form: &yourAddressForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetYourAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetYourAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	address := Address{Line1: "abc"}
	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			You: Person{
				Address: address,
			},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetYourAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
				Action:  "manual",
				Address: &Address{},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetYourAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App:  appData,
			Form: &yourAddressForm{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostYourAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
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

	err := YourAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whoIsTheLpaForPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostYourAddressManualWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
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

	err := YourAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostYourAddressManualFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{
			You: Person{
				FirstNames: "John",
				Address:    Address{Line1: "abc"},
			},
			WhoFor: "me",
		}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
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

	err := YourAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whoIsTheLpaForPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostYourAddressManualWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
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

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostYourAddressSelect(t *testing.T) {
	w := httptest.NewRecorder()

	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			You: Person{
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

	err := YourAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whoIsTheLpaForPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostYourAddressSelectWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	response := ordnance_survey.PostcodeLookupResponse{
		TotalResults: 1,
		Results: []ordnance_survey.AddressDetails{
			{
				Address:           "",
				BuildingName:      "",
				BuildingNumber:    "1",
				ThoroughFareName:  "Road Way",
				DependentLocality: "",
				Town:              "Townville",
				Postcode:          "",
			},
		},
	}

	addresses := []Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(response, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
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

	err := YourAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostYourAddressLookup(t *testing.T) {
	w := httptest.NewRecorder()

	response := ordnance_survey.PostcodeLookupResponse{
		TotalResults: 1,
		Results: []ordnance_survey.AddressDetails{
			{
				Address:           "",
				BuildingName:      "",
				BuildingNumber:    "1",
				ThoroughFareName:  "Road Way",
				DependentLocality: "",
				Town:              "Townville",
				Postcode:          "",
			},
		},
	}

	addresses := []Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(response, nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
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

	err := YourAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostYourAddressLookupError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", "NG1").
		Return(ordnance_survey.PostcodeLookupResponse{TotalResults: 0}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
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

	err := YourAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostYourAddressLookupWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action": {"lookup"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourAddressData{
			App: appData,
			Form: &yourAddressForm{
				Action: "lookup",
			},
			Errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := YourAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadYourAddressForm(t *testing.T) {
	expectedAddress := &Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	testCases := map[string]struct {
		form   url.Values
		result *yourAddressForm
	}{
		"lookup": {
			form: url.Values{
				"action":          {"lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &yourAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"select": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &yourAddressForm{
				Action:  "select",
				Address: expectedAddress,
			},
		},
		"select-not-selected": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {""},
			},
			result: &yourAddressForm{
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
			result: &yourAddressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			actual := readYourAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestYourAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourAddressForm
		errors map[string]string
	}{
		"lookup-valid": {
			form: &yourAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			errors: map[string]string{},
		},
		"lookup-missing-postcode": {
			form: &yourAddressForm{
				Action: "lookup",
			},
			errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		},
		"select-valid": {
			form: &yourAddressForm{
				Action:  "select",
				Address: &Address{},
			},
			errors: map[string]string{},
		},
		"select-not-selected": {
			form: &yourAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: map[string]string{
				"select-address": "selectAddress",
			},
		},
		"manual-valid": {
			form: &yourAddressForm{
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
			form: &yourAddressForm{
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
