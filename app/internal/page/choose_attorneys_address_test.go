package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Form:     &addressForm{},
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Attorneys: actor.Attorneys{attorney},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: address,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App: appData,
			Form: &addressForm{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Form:     &addressForm{},
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
		Return(&Lpa{Attorneys: actor.Attorneys{{
			ID:      "123",
			Address: place.Address{},
		}}}, nil)

	lpaStore.
		On("Put", r.Context(), &Lpa{
			Attorneys: actor.Attorneys{{
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
			Attorneys: actor.Attorneys{{
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
			Attorneys: actor.Attorneys{{
				ID:         "123",
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			}},
			WhoFor: "me",
		}, nil)

	lpaStore.
		On("Put", r.Context(), &Lpa{
			Attorneys: actor.Attorneys{{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

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
			Form: &addressForm{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	updatedAttorney := actor.Attorney{
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
		On("Put", r.Context(), &Lpa{Attorneys: actor.Attorneys{updatedAttorney}}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
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

func TestPostChooseAttorneysInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}
	logger.
		On("Print", invalidPostcodeErr)

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(logger, template.Func, addressClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostChooseAttorneysValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	form := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	logger := &mockLogger{}

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	addressClient := &mockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
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

	attorney := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Attorneys: actor.Attorneys{attorney}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &chooseAttorneysAddressData{
			App:      appData,
			Attorney: attorney,
			Form: &addressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
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
				Attorneys: actor.Attorneys{
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
