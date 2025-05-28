package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChooseAttorneysAddress(t *testing.T) {
	testcases := map[string]string{
		"GB": "",
		"FR": "postcode",
	}

	for country, action := range testcases {
		t.Run(country, func(t *testing.T) {
			uid := actoruid.New()
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

			f := form.NewAddressForm()
			f.Action = action

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &chooseAddressData{
					App:        testAppData,
					Form:       f,
					UID:        uid,
					FullName:   "John Smith",
					ActorLabel: "attorney",
					TitleKeys:  testTitleKeys,
				}).
				Return(nil)

			err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
				Donor: donordata.Donor{Address: place.Address{Line1: "abc", Country: country}},
				Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
					UID:        uid,
					FirstNames: "John",
					LastName:   "Smith",
					Address:    place.Address{},
				}}},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChooseAttorneysAddressFromStore(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	attorney := donordata.Attorney{
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:     donordata.Donor{Address: place.Address{Line1: "abc", Country: "GB"}},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysAddressManual(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?action=manual&id="+uid.String(), nil)

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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChooseAttorneysAddressWhenTemplateErrors(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAddressManual(t *testing.T) {
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

	attorney := donordata.Attorney{
		UID:        uid,
		FirstNames: "a",
		Address:    testAddress,
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(r.Context(), attorney).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:     "lpa-id",
			Tasks:     donordata.Tasks{ChooseAttorneys: task.StateCompleted},
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, nil, nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
			UID:        uid,
			FirstNames: "a",
			Address:    place.Address{},
		}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysAddressManualWhenReuseStoreErrors(t *testing.T) {
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseAttorneysAddress(nil, nil, nil, nil, reuseStore)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
			UID:     uid,
			Address: place.Address{},
		}}},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostChooseAttorneysAddressManualWhenDonorStoreErrors(t *testing.T) {
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(r.Context(), mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ChooseAttorneysAddress(nil, nil, nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
			UID:     uid,
			Address: place.Address{},
		}}},
	})

	assert.Equal(t, expectedError, err)
}

func TestPostChooseAttorneysAddressManualFromStore(t *testing.T) {
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

	attorney := donordata.Attorney{
		UID:        uid,
		FirstNames: "John",
		Address:    testAddress,
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(r.Context(), attorney).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:     "lpa-id",
			Tasks:     donordata.Tasks{ChooseAttorneys: task.StateCompleted},
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, nil, nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{
			UID:        uid,
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysAddressManualWhenValidationError(t *testing.T) {
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

	attorney := donordata.Attorney{
		UID:     uid,
		Address: place.Address{},
	}

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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAddressPostcodeSelect(t *testing.T) {
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAddressPostcodeSelectWhenValidationError(t *testing.T) {
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysPostcodeLookup(t *testing.T) {
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysPostcodeLookupError(t *testing.T) {
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(logger, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysPostcodeLookupInvalidPostcodeError(t *testing.T) {
	w := httptest.NewRecorder()
	invalidPostcodeErr := place.BadRequestError{
		Statuscode: 400,
		Message:    "invalid postcode",
	}

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	uid := actoruid.New()
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(logger, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	uid := actoruid.New()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, addressClient, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysPostcodeLookupWhenValidationError(t *testing.T) {
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
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAddressReuse(t *testing.T) {
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
			ActorLabel: "attorney",
			Addresses:  []place.Address{{Line1: "donor lane", Country: "GB"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:     donordata.Donor{Address: place.Address{Line1: "donor lane", Country: "GB"}},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorney := donordata.Attorney{
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

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		PutAttorney(r.Context(), attorney).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:     "lpa-id",
			Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{attorney}},
			Tasks:     donordata.Tasks{ChooseAttorneys: task.StateInProgress},
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, nil, nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		LpaID:     "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseAttorneysSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChooseAttorneysAddressReuseSelectWhenValidationError(t *testing.T) {
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
			Addresses:  []place.Address{{Line1: "donor lane", Country: "GB"}},
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			UID:        uid,
			FullName:   " ",
			ActorLabel: "attorney",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := ChooseAttorneysAddress(nil, template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:     donordata.Donor{Address: place.Address{Line1: "donor lane", Country: "GB"}},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{UID: uid}}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChooseAttorneysManuallyFromAnotherPage(t *testing.T) {
	uid := actoruid.New()

	testcases := map[string]struct {
		requestUrl      string
		expectedNextUrl string
	}{
		"with from value": {
			"/?from=/lpa/lpa-id/test&id=" + uid.String(),
			"/lpa/lpa-id/test",
		},
		"without from value": {
			"/?from=&id=" + uid.String(),
			donor.PathChooseAttorneysSummary.Format("lpa-id"),
		},
		"missing from key": {
			"/?id=" + uid.String(),
			donor.PathChooseAttorneysSummary.Format("lpa-id"),
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

			attorney := donordata.Attorney{
				UID: uid,
				Address: place.Address{
					Line1:      "a",
					TownOrCity: "b",
					Postcode:   "C",
					Country:    "GB",
				},
			}

			donor := &donordata.Provided{
				LpaID: "lpa-id",
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{attorney},
				},
			}

			reuseStore := newMockReuseStore(t)
			reuseStore.EXPECT().
				PutAttorney(r.Context(), attorney).
				Return(nil)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), donor).
				Return(nil)

			err := ChooseAttorneysAddress(nil, nil, nil, donorStore, reuseStore)(testAppData, w, r, donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedNextUrl, resp.Header.Get("Location"))
		})
	}
}
