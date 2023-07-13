package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProvideAddressToSendEvidenceForm(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:  testAppData,
			Form: &form.AddressForm{},
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideAddressToSendEvidenceFormFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	address := place.Address{Line1: "abc"}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &address,
			},
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		EvidenceFormAddress: address,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideAddressToSendEvidenceFormManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetProvideAddressToSendEvidenceFormWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:  testAppData,
			Form: &form.AddressForm{},
		}).
		Return(expectedError)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormManual(t *testing.T) {
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                  "lpa-id",
			EvidenceFormAddress: testAddress,
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.HowToSendEvidence.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideAddressToSendEvidenceFormManualWhenStoreErrors(t *testing.T) {
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			EvidenceFormAddress: testAddress,
		}).
		Return(expectedError)

	err := ProvideAddressToSendEvidenceForm(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostProvideAddressToSendEvidenceFormManualFromStore(t *testing.T) {
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                  "lpa-id",
			EvidenceFormAddress: testAddress,
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID:                  "lpa-id",
		EvidenceFormAddress: place.Address{Line1: "abc"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.HowToSendEvidence.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostProvideAddressToSendEvidenceFormManualWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "d",
				},
			},
			Errors: validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormSelect(t *testing.T) {
	expectedAddress := &place.Address{
		Line1:      "a",
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {expectedAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        expectedAddress,
			},
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

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
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormLookup(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := newMockAddressClient(t)
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(logger, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.
		On("Print", invalidPostcodeErr)

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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(logger, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(logger, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostProvideAddressToSendEvidenceFormLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "postcode",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := ProvideAddressToSendEvidenceForm(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
