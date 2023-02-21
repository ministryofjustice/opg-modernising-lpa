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

func TestGetChoosePeopleToNotifyAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			Form:           &form.AddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetChoosePeopleToNotifyAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: testAddress,
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &testAddress,
			},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: testAddress,
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			PersonToNotify: personToNotify,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			Form:           &form.AddressForm{},
			PersonToNotify: personToNotify,
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressManual(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskInProgress},
		}, nil)

	personToNotify.Address = testAddress

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualWhenStoreErrors(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	personToNotify.Address = testAddress

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressManualFromStore(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "line1"},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskInProgress},
		}, nil)

	personToNotify.Address = testAddress

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: page.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostChoosePeopleToNotifyAddressSelect(t *testing.T) {
	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		FirstNames: "John",
		Address:    place.Address{Line1: "abc"},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	updatedPersonToNotify := actor.PersonToNotify{
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
		On("Put", r.Context(), &page.Lpa{PeopleToNotify: actor.PeopleToNotify{updatedPersonToNotify}}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &testAddress,
			},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostChoosePeopleToNotifyAddressSelectWhenValidationError(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookup(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:         "123",
		Address:    place.Address{},
		FirstNames: "John",
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookupError(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressInvalidPostcodeError(t *testing.T) {
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

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &choosePeopleToNotifyAddressData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Form: &form.AddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
			"/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
		"missing from key": {
			"/?id=123",
			"/lpa/lpa-id" + page.Paths.ChoosePeopleToNotifySummary,
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			f := url.Values{
				"action":           {"manual"},
				"address-line-1":   {"a"},
				"address-town":     {"b"},
				"address-postcode": {"c"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					PeopleToNotify: actor.PeopleToNotify{
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
				On("Put", r.Context(), &page.Lpa{
					PeopleToNotify: actor.PeopleToNotify{
						{
							ID: "123",
							Address: place.Address{
								Line1:      "a",
								TownOrCity: "b",
								Postcode:   "c",
							},
						},
					},
					Tasks: page.Tasks{PeopleToNotify: page.TaskCompleted},
				}).
				Return(nil)

			err := ChoosePeopleToNotifyAddress(nil, nil, nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
		})
	}
}
