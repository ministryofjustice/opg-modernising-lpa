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

func TestGetChooseReplacementAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Form:     &chooseAttorneysAddressForm{},
			Attorney: ra,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseReplacementAttorneysAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChooseReplacementAttorneysAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChooseReplacementAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App: appData,
			Form: &chooseAttorneysAddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			Attorney: ra,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetChooseReplacementAttorneysAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Form:     &chooseAttorneysAddressForm{},
			Attorney: ra,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseReplacementAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	ra.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			ReplacementAttorneys: []Attorney{ra},
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

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseReplacementAttorneysAddressManualWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	ra.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{ReplacementAttorneys: []Attorney{ra}}).
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

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseReplacementAttorneysAddressManualFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			ReplacementAttorneys: []Attorney{ra},
			WhoFor:               "me",
		}, nil)

	ra.Address = address

	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			ReplacementAttorneys: []Attorney{ra},
			WhoFor:               "me",
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

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/choose-replacement-attorneys-summary", resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChooseReplacementAttorneysAddressManualWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChooseReplacementAttorneysAddressSelect(t *testing.T) {
	w := httptest.NewRecorder()

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	updatedRa := Attorney{
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
		On("Put", mock.Anything, "session-id", &Lpa{ReplacementAttorneys: []Attorney{updatedRa}}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChooseReplacementAttorneysAddressSelectWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
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

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseReplacementAttorneysAddressLookup(t *testing.T) {
	w := httptest.NewRecorder()

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
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

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostChooseReplacementAttorneysAddressLookupError(t *testing.T) {
	w := httptest.NewRecorder()

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form: &chooseAttorneysAddressForm{
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

	err := ChooseReplacementAttorneysAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChooseReplacementAttorneysAddressLookupWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action": {"lookup"},
	}

	ra := Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{ReplacementAttorneys: []Attorney{ra}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseReplacementAttorneysAddressData{
			App:      appData,
			Attorney: ra,
			Form: &chooseAttorneysAddressForm{
				Action: "lookup",
			},
			Errors: map[string]string{
				"lookup-postcode": "enterPostcode",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ChooseReplacementAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostChooseReplacementAttorneysManuallyFromAnotherPage(t *testing.T) {
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
			"/choose-replacement-attorneys-summary",
		},
		"missing from key": {
			"/?id=123",
			"/choose-replacement-attorneys-summary",
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpa := &Lpa{
				ReplacementAttorneys: []Attorney{
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

			err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}
