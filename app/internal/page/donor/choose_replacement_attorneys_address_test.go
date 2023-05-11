package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Form:     &form.AddressForm{},
			Attorney: ra,
			Lpa:      &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: testAddress,
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &testAddress,
			},
			Lpa: &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: testAddress,
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			Attorney: ra,
			Lpa:      &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Form:     &form.AddressForm{},
			Attorney: ra,
			Lpa:      &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressSkip(t *testing.T) {
	f := url.Values{
		"action":           {"skip"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{
				ID:         "123",
				FirstNames: "a",
				Email:      "a",
				Address:    place.Address{Line1: "abc"},
			}},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{ID: "123", FirstNames: "a", Email: "a"}},
			Tasks:                page.Tasks{ChooseReplacementAttorneys: page.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManual(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{ID: "123", FirstNames: "a"}},
		}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{
				ID:         "123",
				FirstNames: "a",
				Address:    testAddress,
			}},
			Tasks: page.Tasks{ChooseReplacementAttorneys: page.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManualWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{{ID: "123"}}}, nil)

	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostChooseReplacementAttorneysAddressManualFromStore(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{
				ID:         "123",
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			}},
			WhoFor: "me",
		}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			ReplacementAttorneys: actor.Attorneys{{
				ID:         "123",
				FirstNames: "John",
				Address:    testAddress,
			}},
			WhoFor: "me",
			Tasks:  page.Tasks{ChooseReplacementAttorneys: page.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			Errors: validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
			Lpa:    &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressSelect(t *testing.T) {
	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	updatedRa := actor.Attorney{
		ID: "123",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "e",
		},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			ReplacementAttorneys: actor.Attorneys{updatedRa},
			Tasks:                page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
		}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &testAddress,
			},
			Lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{updatedRa},
				Tasks:                page.Tasks{ChooseReplacementAttorneys: page.TaskInProgress},
			},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			Lpa:       &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressLookup(t *testing.T) {
	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Lpa:       &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			Lpa:       &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", invalidPostcodeErr)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			Lpa:       &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			Lpa:       &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseReplacementAttorneysAddressData{
			App:      testAppData,
			Attorney: ra,
			Form: &form.AddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			Lpa:    &page.Lpa{ReplacementAttorneys: actor.Attorneys{ra}},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
