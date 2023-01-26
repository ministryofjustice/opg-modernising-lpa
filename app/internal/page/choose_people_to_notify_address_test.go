package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotifyAddress(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &choosePeopleToNotifyAddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App: appData,
			Form: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChoosePeopleToNotifyAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &choosePeopleToNotifyAddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
		}).
		Return(nil)

	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}).
		Return(expectedError)

	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
		}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
		}).
		Return(nil)

	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	personToNotify := PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			Errors: map[string]string{
				"address-line-1": "enterAddress",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChoosePeopleToNotifyAddressSelect(t *testing.T) {
	w := httptest.NewRecorder()

	personToNotify := PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	updatedPersonToNotify := PersonToNotify{
		ID: "123",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "e",
		},
		FirstNames: "John",
	}

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{PeopleToNotify: []PersonToNotify{updatedPersonToNotify}}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &address,
			},
			Errors: map[string]string{},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {address.Encode()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChoosePeopleToNotifyAddressSelectWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors: map[string]string{
				"select-address": "selectAddress",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressLookup(t *testing.T) {
	w := httptest.NewRecorder()

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	personToNotify := PersonToNotify{
		ID:         "123",
		Address:    place.Address{},
		FirstNames: "John",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
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

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChoosePeopleToNotifyAddressLookupError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors: map[string]string{
				"lookup-postcode": "couldNotLookupPostcode",
			},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressNotFoundError(t *testing.T) {
	w := httptest.NewRecorder()
	notFoundErr := place.NotFoundError{
		Statuscode: 400,
		Message:    "not found",
	}

	logger := &mockLogger{}
	logger.
		On("Print", notFoundErr)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, notFoundErr)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors: map[string]string{
				"lookup-postcode": "enterUkPostCode",
			},
		}).
		Return(nil)

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressLookupWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action": {"lookup"},
	}

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action: "lookup",
			},
			Errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadChoosePeopleToNotifyAddressForm(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
	}

	testCases := map[string]struct {
		form   url.Values
		result *choosePeopleToNotifyAddressForm
	}{
		"lookup": {
			form: url.Values{
				"action":          {"lookup"},
				"lookup-postcode": {"NG1"},
			},
			result: &choosePeopleToNotifyAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"select": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {expectedAddress.Encode()},
			},
			result: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: expectedAddress,
			},
		},
		"select not selected": {
			form: url.Values{
				"action":         {"select"},
				"select-address": {""},
			},
			result: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: nil,
			},
		},
		"manual": {
			form: url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-line-2":   {"b"},
				"address-line-3":   {"c"},
				"address-town":     {"d"},
				"address-postcode": {"e"},
			},
			result: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: expectedAddress,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			actual := readChoosePeopleToNotifyAddressForm(r)
			assert.Equal(t, tc.result, actual)
		})
	}
}

func TestChoosePeopleToNotifyAddressFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *choosePeopleToNotifyAddressForm
		errors map[string]string
	}{
		"lookup valid": {
			form: &choosePeopleToNotifyAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			errors: map[string]string{},
		},
		"lookup missing postcode": {
			form: &choosePeopleToNotifyAddressForm{
				Action: "lookup",
			},
			errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		},
		"select valid": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: &place.Address{},
			},
			errors: map[string]string{},
		},
		"select not selected": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: map[string]string{
				"select-address": "selectAddress",
			},
		},
		"manual valid": {
			form: &choosePeopleToNotifyAddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      "a",
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
			errors: map[string]string{},
		},
		"manual missing all": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: map[string]string{
				"address-line-1":   "enterAddress",
				"address-town":     "enterTownOrCity",
				"address-postcode": "enterPostcode",
			},
		},
		"manual max length": {
			form: &choosePeopleToNotifyAddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 50),
					Line2:      strings.Repeat("x", 50),
					Line3:      strings.Repeat("x", 50),
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
			errors: map[string]string{},
		},
		"manual too long": {
			form: &choosePeopleToNotifyAddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      strings.Repeat("x", 51),
					Line2:      strings.Repeat("x", 51),
					Line3:      strings.Repeat("x", 51),
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
			errors: map[string]string{
				"address-line-1": "addressLine1TooLong",
				"address-line-2": "addressLine2TooLong",
				"address-line-3": "addressLine3TooLong",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}

func TestPostPersonToNotifyAddressManuallyFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test&id=123",
			"/test",
		},
		"without from value": {
			"/?from=&id=123",
			appData.Paths.ChoosePeopleToNotifySummary,
		},
		"missing from key": {
			"/?id=123",
			appData.Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpa := &Lpa{
				PeopleToNotify: []PersonToNotify{
					{
						ID: "123",
						Address: place.Address{
							Line1:      "a",
							TownOrCity: "b",
							Postcode:   "c",
						},
					},
				},
			}

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(lpa, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", lpa).
				Return(nil)

			form := url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-town":     {"b"},
				"address-postcode": {"c"},
			}

			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}
