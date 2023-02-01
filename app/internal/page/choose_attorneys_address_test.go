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

func TestGetChooseAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Form:     &chooseAttorneysAddressForm{},
			Attorney: attorney,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseAttorneysAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := ChooseAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseAttorneysAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	attorney := Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Attorneys: []Attorney{attorney},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	attorney := Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			Attorney: attorney,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChooseAttorneysAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Form:     &chooseAttorneysAddressForm{},
			Attorney: attorney,
		}).
		Return(expectedError)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseAttorneysAddressManual(t *testing.T) {
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

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{{
			ID:      "123",
			Address: place.Address{},
		}}}, nil)

	lpaStore.
		On("Put", r.Context(), &Lpa{
			Attorneys: []Attorney{{
				ID:      "123",
				Address: address,
			}},
			Tasks: Tasks{ChooseAttorneys: TaskCompleted},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseAttorneysAddressManualWhenStoreErrors(t *testing.T) {
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

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Attorneys: []Attorney{{
				ID:      "123",
				Address: place.Address{},
			}},
		}, nil)

	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseAttorneysAddressManualFromStore(t *testing.T) {
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

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Attorneys: []Attorney{{
				ID:         "123",
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			}},
			WhoFor: "me",
		}, nil)

	lpaStore.
		On("Put", r.Context(), &Lpa{
			Attorneys: []Attorney{{
				ID:         "123",
				FirstNames: "John",
				Address:    address,
			}},
			WhoFor: "me",
			Tasks:  Tasks{ChooseAttorneys: TaskCompleted},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseAttorneysSummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseAttorneysAddressManualWhenValidationError(t *testing.T) {
	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			Errors: validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChooseAttorneysAddressSelect(t *testing.T) {
	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {address.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	updatedAttorney := Attorney{
		ID: "123",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "e",
		},
	}

	lpaStore.
		On("Put", r.Context(), &Lpa{Attorneys: []Attorney{updatedAttorney}}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &address,
			},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChooseAttorneysAddressSelectWhenValidationError(t *testing.T) {
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

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.AddressSelectedError{Label: "address"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseAttorneysAddressLookup(t *testing.T) {
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

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChooseAttorneysAddressLookupError(t *testing.T) {
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

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChooseAttorneysAddressLookupWhenValidationError(t *testing.T) {
	form := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	attorney := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: []Attorney{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &chooseAttorneysAddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "postcode"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadChooseAttorneysAddressForm(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
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
		"select not selected": {
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
				"address-line-3":   {"c"},
				"address-town":     {"d"},
				"address-postcode": {"e"},
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
		errors validation.List
	}{
		"lookup valid": {
			form: &chooseAttorneysAddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
		},
		"lookup missing postcode": {
			form: &chooseAttorneysAddressForm{
				Action: "lookup",
			},
			errors: validation.With("lookup-postcode", validation.EnterError{Label: "postcode"}),
		},
		"select valid": {
			form: &chooseAttorneysAddressForm{
				Action:  "select",
				Address: &place.Address{},
			},
		},
		"select not selected": {
			form: &chooseAttorneysAddressForm{
				Action:  "select",
				Address: nil,
			},
			errors: validation.With("select-address", validation.AddressSelectedError{Label: "address"}),
		},
		"manual valid": {
			form: &chooseAttorneysAddressForm{
				Action: "manual",
				Address: &place.Address{
					Line1:      "a",
					TownOrCity: "b",
					Postcode:   "c",
				},
			},
		},
		"manual missing all": {
			form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			errors: validation.
				With("address-line-1", validation.EnterError{Label: "addressLine1"}).
				With("address-town", validation.EnterError{Label: "townOrCity"}),
		},
		"manual max length": {
			form: &chooseAttorneysAddressForm{
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
			form: &chooseAttorneysAddressForm{
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

func TestPostChooseAttorneysManuallyFromAnotherPage(t *testing.T) {
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
			"/lpa/lpa-id" + Paths.ChooseAttorneysSummary,
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + Paths.ChooseAttorneysSummary,
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

			lpa := &Lpa{
				Attorneys: []Attorney{
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
				On("Get", r.Context()).
				Return(lpa, nil)
			lpaStore.
				On("Put", r.Context(), lpa).
				Return(nil)

			err := ChooseAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}
