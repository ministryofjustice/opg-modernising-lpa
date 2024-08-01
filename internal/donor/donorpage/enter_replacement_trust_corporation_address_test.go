package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterReplacementTrustCorporationAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{Name: "a"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReplacementTrustCorporationAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "manual",

				FieldNames: form.FieldNames.Address, Address: &place.Address{},
			},
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReplacementTrustCorporationAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressManual(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Tasks: donordata.Tasks{ChooseReplacementAttorneys: task.StateCompleted},
			ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
				Name:    "a",
				Address: testAddress,
			}},
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
			Name:    "a",
			Address: place.Address{},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterReplacementTrustCorporationAddressManualWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterReplacementTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterReplacementTrustCorporationAddressManualFromStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Tasks: donordata.Tasks{ChooseReplacementAttorneys: task.StateCompleted},
			ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
				Name:    "John",
				Address: testAddress,
			}},
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
			Name:    "John",
			Address: place.Address{Line1: "abc"},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterReplacementTrustCorporationAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.TownOrCity: {"c"},
		form.FieldNames.Address.Postcode:   {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invalidAddress := &place.Address{
		Line2:      "b",
		TownOrCity: "c",
		Postcode:   "D",
		Country:    "GB",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "manual",

				FieldNames: form.FieldNames.Address, Address: invalidAddress,
			},
			Errors:     validation.With(form.FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressPostcodeSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-select"},
		"lookup-postcode":              {"NG1"},
		"select-address":               {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        &testAddress,
				FieldNames:     form.FieldNames.Address,
			},
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressPostcodeSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-select"},
		"lookup-postcode":              {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := newMockAddressClient(t)
	addressClient.EXPECT().
		LookupPostcode(mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-select",
				LookupPostcode: "NG1",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses:  addresses,
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationPostcodeLookup(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addresses := []place.Address{
		{Line1: "1 Road Way", TownOrCity: "Townville"},
	}

	addressClient := newMockAddressClient(t)
	addressClient.EXPECT().
		LookupPostcode(mock.Anything, "NG1").
		Return(addresses, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses:  addresses,
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationPostcodeLookupError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "postcode lookup", slog.Any("err", expectedError))

	addressClient := newMockAddressClient(t)
	addressClient.EXPECT().
		LookupPostcode(mock.Anything, "NG1").
		Return([]place.Address{}, expectedError)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "NG1",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressPostcodeLookupInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "postcode lookup", slog.Any("err", invalidPostcodeErr))

	addressClient := newMockAddressClient(t)
	addressClient.EXPECT().
		LookupPostcode(mock.Anything, "XYZ").
		Return([]place.Address{}, invalidPostcodeErr)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "XYZ",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	addressClient := newMockAddressClient(t)
	addressClient.EXPECT().
		LookupPostcode(mock.Anything, "XYZ").
		Return([]place.Address{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "XYZ",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressPostcodeLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "postcode",
				FieldNames: form.FieldNames.Address,
			},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressReuse(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "reuse",
				FieldNames: form.FieldNames.Address,
			},
			ActorLabel: "theTrustCorporation",
			Addresses:  []place.Address{{Line1: "donor lane"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:                donordata.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedTrustCorporation := donordata.TrustCorporation{
		Name: "a",
		Address: place.Address{
			Line1:      "a",
			Line2:      "b",
			Line3:      "c",
			TownOrCity: "d",
			Postcode:   "E",
			Country:    "GB",
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:                "lpa-id",
			ReplacementAttorneys: donordata.Attorneys{TrustCorporation: updatedTrustCorporation},
			Tasks:                donordata.Tasks{ChooseReplacementAttorneys: task.StateCompleted},
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
			Name: "a",
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterReplacementTrustCorporationAddressReuseSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "reuse-select",
				FieldNames: form.FieldNames.Address,
			},
			Addresses:  []place.Address{{Line1: "donor lane"}},
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterReplacementTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:                donordata.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReplacementTrustCorporationAddressManuallyFromAnotherPage(t *testing.T) {
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
			page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"),
		},
		"missing from key": {
			"/",
			page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"),
		},
	}

	for testname, tc := range testcases {
		t.Run(testname, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.Address.Action:     {"manual"},
				form.FieldNames.Address.Line1:      {"a"},
				form.FieldNames.Address.TownOrCity: {"b"},
				form.FieldNames.Address.Postcode:   {"c"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donor := &donordata.Provided{
				LpaID: "lpa-id",
				ReplacementAttorneys: donordata.Attorneys{TrustCorporation: donordata.TrustCorporation{
					Address: place.Address{
						Line1:      "a",
						TownOrCity: "b",
						Postcode:   "c",
					},
				}},
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), donor).
				Return(nil)

			err := EnterReplacementTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
		})
	}
}
