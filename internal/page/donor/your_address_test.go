package donor

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

func TestGetYourAddress(t *testing.T) {
	testcases := map[string]struct {
		query        string
		expectedData *chooseAddressData
	}{
		"first time": {
			expectedData: &chooseAddressData{
				App:       testAppData,
				Form:      form.NewAddressForm(),
				TitleKeys: testTitleKeys,
			},
		},
		"making another LPA": {
			query: "?makingAnotherLPA=1",
			expectedData: &chooseAddressData{
				App:              testAppData,
				Form:             form.NewAddressForm(),
				TitleKeys:        testTitleKeys,
				MakingAnotherLPA: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/"+tc.query, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, tc.expectedData).
				Return(nil)

			err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	address := place.Address{Line1: "abc"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &address,
				FieldNames: form.FieldNames.Address,
			},
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor: actor.Donor{
			Address: address,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAddressManual(t *testing.T) {
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
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:       testAppData,
			Form:      form.NewAddressForm(),
			TitleKeys: testTitleKeys,
		}).
		Return(expectedError)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressManual(t *testing.T) {
	testCases := map[string]struct {
		url              string
		appData          page.AppData
		expectedRedirect string
	}{
		"making first LPA": {
			url:              "/",
			appData:          testAppData,
			expectedRedirect: page.Paths.CanYouSignYourLpa.Format("lpa-id"),
		},
		"making another LPA": {
			url:              "/?makingAnotherLPA=1",
			appData:          testAppData,
			expectedRedirect: page.Paths.WeHaveUpdatedYourDetails.Format("lpa-id") + "?detail=address",
		},
		"supporter": {
			url:              "/",
			appData:          testSupporterAppData,
			expectedRedirect: page.Paths.YourEmail.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.Address.Action:     {"manual"},
				form.FieldNames.Address.Line1:      {"a"},
				form.FieldNames.Address.Line2:      {"b"},
				form.FieldNames.Address.Line3:      {"c"},
				form.FieldNames.Address.TownOrCity: {"d"},
				form.FieldNames.Address.Postcode:   {"e"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &actor.DonorProvidedDetails{
					LpaID: "lpa-id",
					Donor: actor.Donor{
						Address: testAddress,
					},
					HasSentApplicationUpdatedEvent: false,
				}).
				Return(nil)

			err := YourAddress(nil, nil, nil, donorStore)(tc.appData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					Address: place.Address{Line1: "a", Line2: "b", Line3: "c", TownOrCity: "d"},
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourAddressManualWhenAddressNotChanged(t *testing.T) {
	testCases := map[string]struct {
		url              string
		expectedRedirect string
	}{
		"making first LPA": {
			url:              "/",
			expectedRedirect: page.Paths.CanYouSignYourLpa.Format("lpa-id"),
		},
		"making another LPA": {
			url:              "/?makingAnotherLPA=1",
			expectedRedirect: page.Paths.MakeANewLPA.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.Address.Action:     {"manual"},
				form.FieldNames.Address.Line1:      {"a"},
				form.FieldNames.Address.Line2:      {"b"},
				form.FieldNames.Address.Line3:      {"c"},
				form.FieldNames.Address.TownOrCity: {"d"},
				form.FieldNames.Address.Postcode:   {"e"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := YourAddress(nil, nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					Address: place.Address{
						Line1:      "a",
						Line2:      "b",
						Line3:      "c",
						TownOrCity: "d",
						Postcode:   "E",
					},
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourAddressManualWhenStoreErrors(t *testing.T) {
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
			Donor: actor.Donor{
				Address: testAddress,
			},
		}).
		Return(expectedError)

	err := YourAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostYourAddressManualFromStore(t *testing.T) {
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
			Donor: actor.Donor{
				FirstNames: "John",
				Address:    testAddress,
			},
		}).
		Return(nil)

	err := YourAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Donor: actor.Donor{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CanYouSignYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostYourAddressManualWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.TownOrCity: {"c"},
		form.FieldNames.Address.Postcode:   {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action: "manual",
				Address: &place.Address{
					Line2:      "b",
					TownOrCity: "c",
					Postcode:   "D",
					Country:    "GB",
				},
				FieldNames: form.FieldNames.Address,
			},
			Errors:    validation.With(form.FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1OfYourAddress"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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
		form.FieldNames.Address.Action: {"postcode-select"},
		"lookup-postcode":              {"NG1"},
		"select-address":               {expectedAddress.Encode()},
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
				Address:        expectedAddress,
				FieldNames:     form.FieldNames.Address,
			},
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressSelectWhenValidationError(t *testing.T) {
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
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "yourAddressFromTheList"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookup(t *testing.T) {
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
			Addresses: addresses,
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookupError(t *testing.T) {
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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)

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
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noYourAddressesFound"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourAddressLookupWhenValidationError(t *testing.T) {
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
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "yourPostcode"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := YourAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
