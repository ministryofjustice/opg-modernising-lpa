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
		ID:         "123",
		FirstNames: "John",
		LastName:   "Smith",
		Address:    place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			ID:         "123",
			FullName:   "John Smith",
			ActorLabel: "personToNotify",
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: testAddress,
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &testAddress,
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifyAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: testAddress,
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: actor.TaskInProgress},
		}, nil)

	personToNotify.Address = testAddress

	donorStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	personToNotify.Address = testAddress

	donorStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: actor.TaskCompleted},
		}).
		Return(expectedError)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: actor.TaskInProgress},
		}, nil)

	personToNotify.Address = testAddress

	donorStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{personToNotify},
			Tasks:          page.Tasks{PeopleToNotify: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, nil, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChoosePeopleToNotifySummary, resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifyAddressSelect(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{{
				ID:         "123",
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			}},
		}, nil)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			PeopleToNotify: actor.PeopleToNotify{{
				ID: "123",
				Address: place.Address{
					Line1:      "a",
					Line2:      "b",
					Line3:      "c",
					TownOrCity: "d",
					Postcode:   "e",
				},
				FirstNames: "John",
			}},
			Tasks: page.Tasks{PeopleToNotify: actor.TaskCompleted},
		}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &testAddress,
			},
			ID:         "123",
			FullName:   "John ",
			ActorLabel: "personToNotify",
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-select",
				LookupPostcode: "NG1",
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  addresses,
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, addressClient, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookup(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-lookup"},
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
			},
			ID:         "123",
			FullName:   "John ",
			ActorLabel: "personToNotify",
			Addresses:  addresses,
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, addressClient, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-lookup"},
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "NG1",
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, donorStore)(testAppData, w, r)
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
		"action":          {"postcode-lookup"},
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "XYZ",
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "XYZ",
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(logger, template.Execute, addressClient, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChoosePeopleToNotifyAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotify := actor.PersonToNotify{
		ID:      "123",
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{PeopleToNotify: actor.PeopleToNotify{personToNotify}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "postcode",
			},
			ID:         "123",
			FullName:   " ",
			ActorLabel: "personToNotify",
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := ChoosePeopleToNotifyAddress(nil, template.Execute, nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
