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

func TestGetCertificateProviderAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		FirstNames: "John",
		LastName:   "Smith",
		Address:    place.Address{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			FullName:   "John Smith",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{CertificateProvider: certificateProvider})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderAddressWhenProfessionalCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		FirstNames:   "John",
		LastName:     "Smith",
		Address:      place.Address{},
		Relationship: actor.Professionally,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{Action: "postcode"},
			FullName:   "John Smith",
			ActorLabel: "certificateProvider",
			TitleKeys: titleKeys{
				Manual:                          "personsWorkAddress",
				PostcodeSelectAndPostcodeLookup: "selectPersonsWorkAddress",
				Postcode:                        "whatIsPersonsWorkPostcode",
				ReuseAndReuseSelect:             "selectAnAddressForPerson",
				ReuseOrNew:                      "addPersonsAddress",
			},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{CertificateProvider: certificateProvider})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProvider := actor.CertificateProvider{
		Address: testAddress,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &testAddress,
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{CertificateProvider: certificateProvider})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	certificateProvider := actor.CertificateProvider{
		Address: testAddress,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: &place.Address{},
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{CertificateProvider: certificateProvider})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCertificateProviderAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.Lpa{
			ID:                  "lpa-id",
			CertificateProvider: actor.CertificateProvider{Address: testAddress},
			Tasks:               actor.DonorTasks{CertificateProvider: actor.TaskCompleted},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.Lpa{
			CertificateProvider: actor.CertificateProvider{Address: testAddress},
			Tasks:               actor.DonorTasks{CertificateProvider: actor.TaskCompleted},
		}).
		Return(expectedError)

	err := CertificateProviderAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.Lpa{})

	assert.Equal(t, expectedError, err)
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.Lpa{
			ID: "lpa-id",
			CertificateProvider: actor.CertificateProvider{
				FirstNames: "John",
				Address:    testAddress,
			},
			Tasks: actor.DonorTasks{CertificateProvider: actor.TaskCompleted},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.Lpa{
		ID: "lpa-id",
		CertificateProvider: actor.CertificateProvider{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
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

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "D",
		Country:    "GB",
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Errors:     validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeSelect(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &testAddress,
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeSelectWhenValidationError(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  addresses,
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeLookup(t *testing.T) {
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

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  addresses,
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeLookupError(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeLookupInvalidPostcodeError(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderPostcodeLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "postcode",
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
func TestPostCertificateProviderAddressReuse(t *testing.T) {
	f := url.Values{
		"action": {"reuse"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "reuse",
			},
			FullName:   " ",
			ActorLabel: "certificateProvider",
			Addresses:  []place.Address{{Line1: "donor lane"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{
		Donor: actor.Donor{Address: place.Address{Line1: "donor lane"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCertificateProviderAddressReuseSelect(t *testing.T) {
	f := url.Values{
		"action":         {"reuse-select"},
		"select-address": {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &actor.Lpa{
			ID: "lpa-id",
			CertificateProvider: actor.CertificateProvider{
				Address: place.Address{
					Line1:      "a",
					Line2:      "b",
					Line3:      "c",
					TownOrCity: "d",
					Postcode:   "E",
					Country:    "GB",
				},
			},
			Tasks: actor.DonorTasks{CertificateProvider: actor.TaskCompleted},
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostCertificateProviderAddressReuseSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"reuse-select"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "reuse-select",
			},
			Addresses:  []place.Address{{Line1: "donor lane"}},
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			FullName:   " ",
			ActorLabel: "certificateProvider",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := CertificateProviderAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.Lpa{
		Donor: actor.Donor{Address: place.Address{Line1: "donor lane"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
