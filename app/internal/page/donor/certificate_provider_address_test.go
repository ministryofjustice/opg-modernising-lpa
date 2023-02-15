package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		Address: place.Address{},
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{CertificateProvider: certificateProvider}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App:                 TestAppData,
			Form:                &form.AddressForm{},
			CertificateProvider: certificateProvider,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetCertificateProviderAddressWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, ExpectedError)

	err := CertificateProviderAddress(nil, nil, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCertificateProviderAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		Address: TestAddress,
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider: certificateProvider,
		}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App:                 TestAppData,
			CertificateProvider: certificateProvider,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &TestAddress,
			},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestGetCertificateProviderAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	certificateProvider := actor.CertificateProvider{
		Address: TestAddress,
	}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{CertificateProvider: certificateProvider}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			CertificateProvider: certificateProvider,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetCertificateProviderAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App:  TestAppData,
			Form: &form.AddressForm{},
		}).
		Return(ExpectedError)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostCertificateProviderAddressManual(t *testing.T) {
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{Address: TestAddress},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, nil, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.HowDoYouKnowYourCertificateProvider, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCertificateProviderAddressManualWhenStoreErrors(t *testing.T) {
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{Address: TestAddress},
		}).
		Return(ExpectedError)

	err := CertificateProviderAddress(nil, nil, nil, lpaStore)(TestAppData, w, r)

	assert.Equal(t, ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCertificateProviderAddressManualFromStore(t *testing.T) {
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				FirstNames: "John",
				Address:    place.Address{Line1: "abc"},
			},
			WhoFor: "me",
		}, nil)

	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				FirstNames: "John",
				Address:    TestAddress,
			},
			WhoFor: "me",
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, nil, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.HowDoYouKnowYourCertificateProvider, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCertificateProviderAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			Errors: validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostCertificateProviderAddressSelect(t *testing.T) {
	f := url.Values{
		"action":          {"select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {TestAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{Address: TestAddress},
		}).
		Return(nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &TestAddress,
			},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostCertificateProviderAddressSelectWhenValidationError(t *testing.T) {
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

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := &MockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "select",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, addressClient, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostCertificateProviderAddressLookup(t *testing.T) {
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

	addressClient := &MockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return(addresses, nil)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: addresses,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, addressClient, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template)
}

func TestPostCertificateProviderAddressLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := &MockLogger{}
	logger.
		On("Print", ExpectedError)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := &MockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "NG1").
		Return([]place.Address{}, ExpectedError)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "NG1",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{"couldNotLookupPostcode"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Func, addressClient, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostCertificateProviderAddressInvalidPostcodeError(t *testing.T) {
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

	logger := &MockLogger{}
	logger.
		On("Print", invalidPostcodeErr)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := &MockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{"invalidPostcode"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Func, addressClient, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostCertificateProviderAddressValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := &MockLogger{}

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	addressClient := &MockAddressClient{}
	addressClient.
		On("LookupPostcode", mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action:         "lookup",
				LookupPostcode: "XYZ",
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Func, addressClient, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, addressClient, template, logger)
}

func TestPostCertificateProviderAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &MockTemplate{}
	template.
		On("Func", w, &certificateProviderAddressData{
			App: TestAppData,
			Form: &form.AddressForm{
				Action: "lookup",
			},
			Errors: validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Func, nil, lpaStore)(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}
