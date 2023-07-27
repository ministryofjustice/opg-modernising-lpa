package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterTrustCorporationAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterTrustCorporationAddressManual(t *testing.T) {
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
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterTrustCorporationAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App:        testAppData,
			Form:       &form.AddressForm{},
			ActorLabel: "theTrustCorporation",
		}).
		Return(expectedError)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressManual(t *testing.T) {
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
			ID:    "lpa-id",
			Tasks: page.Tasks{ChooseAttorneys: actor.TaskCompleted},
			Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name:    "a",
				Address: testAddress,
			}},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
			Name:    "a",
			Address: place.Address{},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationAddressManualWhenStoreErrors(t *testing.T) {
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
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterTrustCorporationAddressManualFromStore(t *testing.T) {
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
			ID:    "lpa-id",
			Tasks: page.Tasks{ChooseAttorneys: actor.TaskCompleted},
			Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name:    "John",
				Address: testAddress,
			}},
			WhoFor: "me",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
			Name:    "John",
			Address: place.Address{Line1: "abc"},
		}},
		WhoFor: "me",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		"action":           {"manual"},
		"address-line-2":   {"b"},
		"address-town":     {"c"},
		"address-postcode": {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "d",
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:  "manual",
				Address: invalidAddress,
			},
			Errors:     validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeSelect(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
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
				Address:        &testAddress,
			},
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeSelectWhenValidationError(t *testing.T) {
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
			Addresses:  addresses,
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationPostcodeLookup(t *testing.T) {
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
			Addresses:  addresses,
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationPostcodeLookupError(t *testing.T) {
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
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeLookupInvalidPostcodeError(t *testing.T) {
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
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

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
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeLookupWhenValidationError(t *testing.T) {
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
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressReuse(t *testing.T) {
	f := url.Values{
		"action": {"reuse"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "reuse",
			},
			ActorLabel: "theTrustCorporation",
			Addresses:  []place.Address{{Line1: "donor lane"}},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Donor:     actor.Donor{Address: place.Address{Line1: "donor lane"}},
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressReuseSelect(t *testing.T) {
	f := url.Values{
		"action":         {"reuse-select"},
		"select-address": {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedTrustCorporation := actor.TrustCorporation{
		Name: "a",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "e",
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:        "lpa-id",
			Attorneys: actor.Attorneys{TrustCorporation: updatedTrustCorporation},
			Tasks:     page.Tasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID: "lpa-id",
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
			Name: "a",
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationAddressReuseSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"reuse-select"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
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
			ActorLabel: "theTrustCorporation",
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Donor:     actor.Donor{Address: place.Address{Line1: "donor lane"}},
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressManuallyFromAnotherPage(t *testing.T) {
	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/test",
			"/test",
		},
		"without from value": {
			"/?from=",
			page.Paths.ChooseAttorneysSummary.Format("lpa-id"),
		},
		"missing from key": {
			"/",
			page.Paths.ChooseAttorneysSummary.Format("lpa-id"),
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

			lpa := &page.Lpa{
				ID: "lpa-id",
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
					Address: place.Address{
						Line1:      "a",
						TownOrCity: "b",
						Postcode:   "c",
					},
				}},
			}

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), lpa).
				Return(nil)

			err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
		})
	}
}
