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

func TestGetEnterPersonToNotifyAddress(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	personToNotify := donordata.PersonToNotify{
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
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotify}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterPersonToNotifyAddressFromStore(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

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
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		PeopleToNotify: donordata.PeopleToNotify{{UID: uid, Address: testAddress}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterPersonToNotifyAddressManual(t *testing.T) {
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
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid, Address: testAddress}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterPersonToNotifyAddressWhenTemplateErrors(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	personToNotify := donordata.PersonToNotify{
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
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(expectedError)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotify}})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyAddressManual(t *testing.T) {
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

	personToNotify := donordata.PersonToNotify{UID: uid, Address: testAddress}

	provided := &donordata.Provided{
		LpaID:          "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{{UID: uid}},
		Tasks:          donordata.Tasks{PeopleToNotify: task.StateInProgress},
	}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Put(r.Context(), provided, personToNotify).
		Return(uid, nil)

	err := EnterPersonToNotifyAddress(nil, nil, nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterPersonToNotifyAddressManualWhenServiceErrors(t *testing.T) {
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

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Put(mock.Anything, mock.Anything, mock.Anything).
		Return(uid, expectedError)

	err := EnterPersonToNotifyAddress(nil, nil, nil, service)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}})

	assert.Equal(t, expectedError, err)
}

func TestPostEnterPersonToNotifyAddressManualFromStore(t *testing.T) {
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

	personToNotify := donordata.PersonToNotify{
		UID:        uid,
		FirstNames: "John",
		Address:    testAddress,
	}

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{donordata.PersonToNotify{
			UID:        uid,
			FirstNames: "John",
			Address:    place.Address{Line1: "line1"},
		}},
		Tasks: donordata.Tasks{PeopleToNotify: task.StateInProgress},
	}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Put(r.Context(), provided, personToNotify).
		Return(uid, nil)

	err := EnterPersonToNotifyAddress(nil, nil, nil, service)(testAppData, w, r, provided)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterPersonToNotifyPostcodeSelect(t *testing.T) {
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
			FullName:   "John ",
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		PeopleToNotify: donordata.PeopleToNotify{{
			UID:        uid,
			FirstNames: "John",
			Address:    place.Address{Line1: "abc"},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeSelectWhenValidationError(t *testing.T) {
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  addresses,
			Errors:     validation.With("select-address", validation.SelectError{Label: "anAddressFromTheList"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeLookup(t *testing.T) {
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
			UID:        uid,
			FullName:   "John ",
			ActorLabel: "personToNotify",
			Addresses:  addresses,
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid, FirstNames: "John"}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeLookupError(t *testing.T) {
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "couldNotLookupPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeLookupInvalidPostcodeError(t *testing.T) {
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "invalidPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeLookupValidPostcodeNoAddresses(t *testing.T) {
	w := httptest.NewRecorder()

	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
		"lookup-postcode":              {"XYZ"},
	}

	uid := actoruid.New()
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
			UID:        uid,
			FullName:   " ",
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{},
			Errors:     validation.With("lookup-postcode", validation.CustomError{Label: "noAddressesFound"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(logger, template.Execute, addressClient, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyPostcodeLookupWhenValidationError(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"postcode-lookup"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotify := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &chooseAddressData{
			App: testAppData,
			Form: &form.AddressForm{
				Action:     "postcode",
				FieldNames: form.FieldNames.Address,
			},
			UID:        uid,
			FullName:   " ",
			ActorLabel: "personToNotify",
			Errors:     validation.With("lookup-postcode", validation.EnterError{Label: "aPostcode"}),
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotify}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyAddressReuse(t *testing.T) {
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
			ActorLabel: "personToNotify",
			Addresses:  []place.Address{{Line1: "donor lane", Country: "GB"}},
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:          donordata.Donor{Address: place.Address{Line1: "donor lane", Country: "GB"}},
		PeopleToNotify: donordata.PeopleToNotify{{UID: uid}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterPersonToNotifyAddressReuseSelect(t *testing.T) {
	f := url.Values{
		form.FieldNames.Address.Action: {"reuse-select"},
		"select-address":               {testAddress.Encode()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotify := donordata.PersonToNotify{
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

	provided := &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{{UID: uid}}}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Put(r.Context(), provided, personToNotify).
		Return(uid, nil)

	err := EnterPersonToNotifyAddress(nil, nil, nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterPersonToNotifyAddressReuseSelectWhenValidationError(t *testing.T) {
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
			ActorLabel: "personToNotify",
			TitleKeys:  testTitleKeys,
		}).
		Return(nil)

	err := EnterPersonToNotifyAddress(nil, template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor:          donordata.Donor{Address: place.Address{Line1: "donor lane", Country: "GB"}},
		PeopleToNotify: donordata.PeopleToNotify{{UID: uid}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
