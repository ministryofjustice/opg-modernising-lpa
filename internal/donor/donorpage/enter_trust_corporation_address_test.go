package donorpage

import (
	"log/slog"
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

func TestGetEnterTrustCorporationAddress(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &place.Address{},
				FieldNames: form.FieldNames.Address,
			},
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressManual(t *testing.T) {
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
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Tasks: actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
			Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name:    "a",
				Address: testAddress,
			}},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
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

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterTrustCorporationAddressManualFromStore(t *testing.T) {
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
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			Tasks: actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
			Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name:    "John",
				Address: testAddress,
			}},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
			Name:    "John",
			Address: place.Address{Line1: "abc"},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterTrustCorporationAddressManualWhenValidationError(t *testing.T) {
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
				Action:     "manual",
				Address:    invalidAddress,
				FieldNames: form.FieldNames.Address,
			},
			Errors:     validation.With(form.FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1"}),
			ActorLabel: "theTrustCorporation",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeSelect(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeSelectWhenValidationError(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationPostcodeLookup(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationPostcodeLookupError(t *testing.T) {
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

	err := EnterTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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

	err := EnterTrustCorporationAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressPostcodeLookupWhenValidationError(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressReuse(t *testing.T) {
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

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor:     actor.Donor{Address: place.Address{Line1: "donor lane"}},
		Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterTrustCorporationAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
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
			Postcode:   "E",
			Country:    "GB",
		},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:     "lpa-id",
			Attorneys: actor.Attorneys{TrustCorporation: updatedTrustCorporation},
			Tasks:     actor.DonorTasks{ChooseAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
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

	err := EnterTrustCorporationAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
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
				form.FieldNames.Address.Action:     {"manual"},
				form.FieldNames.Address.Line1:      {"a"},
				form.FieldNames.Address.TownOrCity: {"b"},
				form.FieldNames.Address.Postcode:   {"c"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.requestUrl, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donor := &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
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

			err := EnterTrustCorporationAddress(nil, nil, nil, donorStore)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
		})
	}
}
