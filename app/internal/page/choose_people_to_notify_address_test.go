package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotifyAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &choosePeopleToNotifyAddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChoosePeopleToNotifyAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			Form:           &choosePeopleToNotifyAddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressManual(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
			Tasks:          Tasks{PeopleToNotify: TaskInProgress},
		}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
			Tasks:          Tasks{PeopleToNotify: TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
			Tasks:          Tasks{PeopleToNotify: TaskCompleted},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualFromStore(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "line1"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
			Tasks:          Tasks{PeopleToNotify: TaskInProgress},
		}, nil)

	personToNotify.Address = address

	lpaStore.
		On("Put", r.Context(), &Lpa{
			PeopleToNotify: []PersonToNotify{personToNotify},
			Tasks:          Tasks{PeopleToNotify: TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressSelect(t *testing.T) {
	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {address.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
		On("Put", r.Context(), &Lpa{PeopleToNotify: []PersonToNotify{updatedPersonToNotify}}).
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
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChoosePeopleToNotifyAddressSelectWhenValidationError(t *testing.T) {
	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
			Errors:    validation.With("select-address", validation.SelectError{Label: "address"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChoosePeopleToNotifyAddressLookup(t *testing.T) {
	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

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
		On("Get", r.Context()).
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
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChoosePeopleToNotifyAddressLookupError(t *testing.T) {
	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
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
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChoosePeopleToNotifyAddressLookupWhenValidationError(t *testing.T) {
	form := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	personToNotify := PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{PeopleToNotify: []PersonToNotify{personToNotify}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &choosePeopleToNotifyAddressData{
			App:            appData,
			PersonToNotify: personToNotify,
			Form: &choosePeopleToNotifyAddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "postcode"}),
		}).
		Return(nil)

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
		errors validation.List
	}{
		"lookup valid": {
			form: &choosePeopleToNotifyAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"lookup missing postcode": {
			form: &choosePeopleToNotifyAddressForm{
				Action: "lookup",
			},
			errors: validation.With("lookup-postcode", validation.EnterError{Label: "postcode"}),
		},
		"select valid": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: &place.Address{},
			},
		},
		"select not selected": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: validation.With("select-address", validation.SelectError{Label: "address"}),
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
		},
		"manual missing all": {
			form: &choosePeopleToNotifyAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With("address-line-1", validation.EnterError{Label: "addressLine1"}).
				With("address-town", validation.EnterError{Label: "townOrCity"}),
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
			errors: validation.
				With("address-line-1", validation.StringTooLongError{Label: "addressLine1", Length: 50}).
				With("address-line-2", validation.StringTooLongError{Label: "addressLine2Label", Length: 50}).
				With("address-line-3", validation.StringTooLongError{Label: "addressLine3Label", Length: 50}),
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
			"/lpa/lpa-id/test",
		},
		"without from value": {
			"/?from=&id=123",
			"/lpa/lpa-id" + Paths.ChoosePeopleToNotifySummary,
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			form := url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-town":     {"b"},
				"address-postcode": {"c"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
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
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
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
					Tasks: Tasks{PeopleToNotify: TaskCompleted},
				}).
				Return(nil)

			err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}
