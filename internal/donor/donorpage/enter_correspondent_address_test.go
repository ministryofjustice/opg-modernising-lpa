package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterCorrespondentAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   "John Smith",
			Form:       form.NewAddressForm(),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Correspondent: donordata.Correspondent{
			FirstNames: "John",
			LastName:   "Smith",
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterCorrespondentAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	address := place.Address{Line1: "abc"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &address,
				FieldNames: form.FieldNames.Address,
			},
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Correspondent: donordata.Correspondent{
			Address: address,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterCorrespondentAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &place.Address{},
				FieldNames: form.FieldNames.Address,
			},
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterCorrespondentAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form:       form.NewAddressForm(),
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressManual(t *testing.T) {
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
			LpaID:  "lpa-id",
			LpaUID: "lpa-uid",
			Correspondent: donordata.Correspondent{
				Address: testAddress,
			},
			Tasks: donordata.Tasks{
				AddCorrespondent: task.StateCompleted,
			},
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(r.Context(), event.CorrespondentUpdated{
			UID:     "lpa-uid",
			Address: &testAddress,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, nil, nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterCorrespondentAddressManualWhenEventClientErrors(t *testing.T) {
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterCorrespondentAddress(nil, nil, nil, nil, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
	})
	assert.Equal(t, expectedError, err)
}

func TestPostEnterCorrespondentAddressManualWhenStoreErrors(t *testing.T) {
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(nil)

	err := EnterCorrespondentAddress(nil, nil, nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterCorrespondentAddressManualFromStore(t *testing.T) {
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
			Correspondent: donordata.Correspondent{
				FirstNames: "John",
				Address:    testAddress,
			},
			Tasks: donordata.Tasks{
				AddCorrespondent: task.StateCompleted,
			},
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(nil)

	err := EnterCorrespondentAddress(nil, nil, nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Correspondent: donordata.Correspondent{
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterCorrespondentAddressManualWhenValidationError(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
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
			Errors:    validation.With(form.FieldNames.Address.Line1, validation.EnterError{Label: "addressLine1"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressSelect(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:         "manual",
				LookupPostcode: "NG1",
				Address:        expectedAddress,
				FieldNames:     form.FieldNames.Address,
			},
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressSelectWhenValidationError(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:         "postcode-select",
				LookupPostcode: "NG1",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses: addresses,
			Errors:    validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressLookup(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:         "postcode-lookup",
				LookupPostcode: "NG1",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses: addresses,
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressLookupError(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
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

	err := EnterCorrespondentAddress(logger, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressInvalidPostcodeError(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
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

	err := EnterCorrespondentAddress(logger, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressValidPostcodeNoAddresses(t *testing.T) {
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
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:         "postcode",
				LookupPostcode: "XYZ",
				FieldNames:     form.FieldNames.Address,
			},
			Addresses: []place.Address{},
			Errors:    validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(logger, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			ActorLabel: "correspondent",
			FullName:   " ",
			Form: &form.AddressForm{
				Action:     "postcode",
				FieldNames: form.FieldNames.Address,
			},
			Errors:    validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			TitleKeys: testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressReuse(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "correspondent",
			Addresses:  []place.Address{{Line1: "donor lane"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{Address: place.Address{Line1: "donor lane"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterCorrespondentAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedCorrespondent := donordata.Correspondent{
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
			LpaID:         "lpa-id",
			Correspondent: updatedCorrespondent,
			Tasks:         donordata.Tasks{AddCorrespondent: task.StateCompleted},
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(nil)

	err := EnterCorrespondentAddress(nil, nil, nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterCorrespondentAddressReuseSelectWhenError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(nil)

	err := EnterCorrespondentAddress(nil, nil, nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
	})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterCorrespondentAddressReuseSelectWhenValidationError(t *testing.T) {
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
			FullName:   " ",
			ActorLabel: "correspondent",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterCorrespondentAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{Address: place.Address{Line1: "donor lane"}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
