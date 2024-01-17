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

func TestGetChooseReplacementAttorneysAddress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:         "123",
		FirstNames: "John",
		LastName:   "Smith",
		Address:    place.Address{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			ID:         "123",
			FullName:   "John Smith",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: testAddress,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &testAddress,
				FieldNames: form.FieldNames.Address,
			},
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressManual(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: testAddress,
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    &place.Address{},
				FieldNames: form.FieldNames.Address,
			},
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=123", nil)

	ra := actor.Attorney{
		ID:      "123",
		Address: place.Address{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{ra}}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressSkip(t *testing.T) {
	f := url.Values{
		"action":           {"skip"},
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
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:                "lpa-id",
			ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123", FirstNames: "a", Email: "a"}}},
			Tasks:                actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
			ID:         "123",
			FirstNames: "a",
			Email:      "a",
			Address:    place.Address{Line1: "abc"},
		}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManual(t *testing.T) {
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
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				ID:         "123",
				FirstNames: "a",
				Address:    testAddress,
			}}},
			Tasks: actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                "lpa-id",
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123", FirstNames: "a"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManualWhenStoreErrors(t *testing.T) {
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
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})

	assert.Equal(t, expectedError, err)
}

func TestPostChooseReplacementAttorneysAddressManualFromStore(t *testing.T) {
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
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID: "lpa-id",
			ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				ID:         "123",
				FirstNames: "John",
				Address:    testAddress,
			}}},
			Tasks: actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
			ID:         "123",
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManualWhenValidationError(t *testing.T) {
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
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "manual",
				Address:    invalidAddress,
				FieldNames: form.FieldNames.Address,
			},
			Errors:     validation.With("address-line-1", validation.EnterError{Label: "addressLine1"}),
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeSelect(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-select"},
		"lookup-postcode": {"NG1"},
		"select-address":  {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeSelectWhenValidationError(t *testing.T) {
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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookup(t *testing.T) {
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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupError(t *testing.T) {
	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"NG1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	logger := newMockLogger(t)
	logger.EXPECT().
		Print(expectedError)

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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupInvalidPostcodeError(t *testing.T) {
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
	logger.EXPECT().
		Print(invalidPostcodeErr)

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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		"action":          {"postcode-lookup"},
		"lookup-postcode": {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
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
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"postcode-lookup"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressReuse(t *testing.T) {
	f := url.Values{
		"action": {"reuse"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "reuse",
				FieldNames: form.FieldNames.Address,
			},
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			Addresses:  []place.Address{{Line1: "donor lane"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor:                actor.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressReuseSelect(t *testing.T) {
	f := url.Values{
		"action":         {"reuse-select"},
		"select-address": {testAddress.Encode()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedAttorney := actor.Attorney{
		ID: "123",
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
			LpaID:                "lpa-id",
			ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{updatedAttorney}},
			Tasks:                actor.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:                "lpa-id",
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressReuseSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		"action": {"reuse-select"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=123", strings.NewReader(f.Encode()))
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
			ID:         "123",
			FullName:   " ",
			CanSkip:    true,
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		Donor:                actor.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "123"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
