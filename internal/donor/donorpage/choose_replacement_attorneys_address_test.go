package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseReplacementAttorneysAddress(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	ra := donordata.Attorney{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
		Address:    place.Address{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			UID:        uid,
			FullName:   "John Smith",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressFromStore(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	ra := donordata.Attorney{
		UID:     uid,
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressManual(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id="+uid.String(), nil)

	ra := donordata.Attorney{
		UID:     uid,
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{ra}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseReplacementAttorneysAddressWhenTemplateErrors(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	ra := donordata.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App:        testAppData,
			Form:       form.NewAddressForm(),
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{ra}}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressManual(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID: "lpa-id",
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
				UID:        uid,
				FirstNames: "a",
				Address:    testAddress,
			}}},
			Tasks: donordata.DonorTasks{ChooseReplacementAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:                "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid, FirstNames: "a"}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressManualWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})

	assert.Equal(t, expectedError, err)
}

func TestPostChooseReplacementAttorneysAddressManualFromStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line1:      {"a"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.Line3:      {"c"},
		form.FieldNames.Address.TownOrCity: {"d"},
		form.FieldNames.Address.Postcode:   {"e"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID: "lpa-id",
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
				UID:        uid,
				FirstNames: "John",
				Address:    testAddress,
			}}},
			Tasks: donordata.DonorTasks{ChooseReplacementAttorneys: actor.TaskCompleted},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID: "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
			UID:        uid,
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
		form.FieldNames.Address.Action:     {"manual"},
		form.FieldNames.Address.Line2:      {"b"},
		form.FieldNames.Address.TownOrCity: {"c"},
		form.FieldNames.Address.Postcode:   {"d"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-select"},
		"lookup-postcode":              {"NG1"},
		"select-address":               {testAddress.Encode()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-select"},
		"lookup-postcode":              {"NG1"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookup(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"NG1"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"NG1"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupInvalidPostcodeError(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysPostcodeLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressReuse(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "reuse",
				FieldNames: form.FieldNames.Address,
			},
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			Addresses:  []place.Address{{Line1: "donor lane"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		Donor:                donordata.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseReplacementAttorneysAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	updatedAttorney := donordata.Attorney{
		UID: uid,
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
		Put(r.Context(), &donordata.DonorProvidedDetails{
			LpaID:                "lpa-id",
			ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{updatedAttorney}},
			Tasks:                donordata.DonorTasks{ChooseReplacementAttorneys: actor.TaskInProgress},
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, nil, nil, donorStore)(testAppData, w, r, &donordata.DonorProvidedDetails{
		LpaID:                "lpa-id",
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChooseReplacementAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseReplacementAttorneysAddressReuseSelectWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "replacementAttorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseReplacementAttorneysAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.DonorProvidedDetails{
		Donor:                donordata.Donor{Address: place.Address{Line1: "donor lane"}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
