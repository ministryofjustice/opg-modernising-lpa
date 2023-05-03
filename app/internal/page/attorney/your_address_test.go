package attorney

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

var testAddress = place.Address{
	Line1:      "a",
	Line2:      "b",
	Line3:      "c",
	TownOrCity: "d",
	Postcode:   "e",
}

func TestGetYourAddress(t *testing.T) {
	testcases := map[string]struct {
		lpa     *page.Lpa
		appData page.AppData
	}{
		"attorney": {
			lpa: &page.Lpa{
				ID:                      "lpa-id",
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			appData: testAppData,
		},
		"replacement attorney": {
			lpa: &page.Lpa{
				ID:                                 "lpa-id",
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &yourAddressData{
					App:  tc.appData,
					Form: &form.AddressForm{},
				}).
				Return(nil)

			err := YourAddress(nil, template.Execute, nil, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := YourAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAddressFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData page.AppData
		lpa     *page.Lpa
	}{
		"attorney": {
			appData: testAppData,
			lpa: &page.Lpa{
				Attorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Address: testAddress,
					},
				},
			},
		},
		"replacement attorney": {
			appData: testReplacementAppData,
			lpa: &page.Lpa{
				ReplacementAttorneys: actor.Attorneys{{ID: "attorney-id", FirstNames: "Bob", LastName: "Smith"}},
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Address: testAddress,
					},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &yourAddressData{
					App: tc.appData,
					Form: &form.AddressForm{
						Action:  "manual",
						Address: &testAddress,
					},
				}).
				Return(nil)

			err := YourAddress(nil, template.Execute, nil, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App:  testAppData,
			Form: &form.AddressForm{},
		}).
		Return(expectedError)

	err := YourAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressManual(t *testing.T) {
	testCases := map[string]struct {
		lpa        *page.Lpa
		updatedLpa *page.Lpa
		appData    page.AppData
	}{
		"attorney": {
			lpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				AttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Address: testAddress,
					},
				},
				AttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testAppData,
		},
		"replacement attorney": {
			lpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{"attorney-id": {}},
			},
			updatedLpa: &page.Lpa{
				ReplacementAttorneyProvidedDetails: map[string]actor.AttorneyProvidedDetails{
					"attorney-id": {
						Address: testAddress,
					},
				},
				ReplacementAttorneyTasks: map[string]page.AttorneyTasks{
					"attorney-id": {
						ConfirmYourDetails: page.TaskCompleted,
					},
				},
			},
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-line-2":   {"b"},
				"address-line-3":   {"c"},
				"address-town":     {"d"},
				"address-postcode": {"e"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(tc.lpa, nil)
			lpaStore.
				On("Put", r.Context(), tc.updatedLpa).
				Return(nil)

			err := YourAddress(nil, nil, nil, lpaStore)(tc.appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.ReadTheLpa, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourAddressManualWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-1":   {"a"},
		"address-line-2":   {"b"},
		"address-line-3":   {"c"},
		"address-town":     {"d"},
		"address-postcode": {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := YourAddress(nil, nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostYourAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
			Errors: validation.With("address-line-1", validation.EnterError{Label: "addressLine1OfYourAddress"}),
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressSelect(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {expectedAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        expectedAddress,
			},
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "yourAddressFromTheList"}),
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookup(t *testing.T) {
	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", invalidPostcodeErr)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"}),
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "yourPostcode"}),
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
