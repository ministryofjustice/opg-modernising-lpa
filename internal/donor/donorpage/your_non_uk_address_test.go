package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourNonUKAddress(t *testing.T) {
	testcases := map[string]struct {
		query        string
		expectedData *yourNonUKAddressData
	}{
		"first time": {
			expectedData: &yourNonUKAddressData{
				App:             testAppData,
				Form:            &yourNonUKAddressForm{},
				WhatCountryLink: donor.PathWhatCountryDoYouLiveIn.Format("lpa-id"),
			},
		},
		"making another LPA": {
			query: "?makingAnotherLPA=1",
			expectedData: &yourNonUKAddressData{
				App:              testAppData,
				Form:             &yourNonUKAddressForm{},
				MakingAnotherLPA: true,
				WhatCountryLink: donor.PathWhatCountryDoYouLiveIn.FormatQuery("lpa-id", url.Values{
					"makingAnotherLPA": {"1"},
				}),
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

			err := YourNonUKAddress(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourNonUKAddressFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	address := place.InternationalAddress{BuildingNumber: "abc", Country: "DE"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourNonUKAddressData{
			App: testAppData,
			Form: &yourNonUKAddressForm{
				Address: address,
			},
			Country:         "DE",
			WhatCountryLink: donor.PathWhatCountryDoYouLiveIn.Format("lpa-id"),
		}).
		Return(nil)

	err := YourNonUKAddress(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			InternationalAddress: address,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourNonUKAddressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := YourNonUKAddress(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourNonUKAddress(t *testing.T) {
	testCases := map[string]struct {
		url              string
		appData          appcontext.Data
		expectedRedirect string
	}{
		"making first LPA": {
			url:              "/",
			appData:          testAppData,
			expectedRedirect: donor.PathReceivingUpdatesAboutYourLpa.Format("lpa-id"),
		},
		"making another LPA": {
			url:              "/?makingAnotherLPA=1",
			appData:          testAppData,
			expectedRedirect: donor.PathWeHaveUpdatedYourDetails.Format("lpa-id") + "?detail=address",
		},
		"supporter": {
			url:              "/",
			appData:          testSupporterAppData,
			expectedRedirect: donor.PathYourEmail.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"apartmentNumber": {"a"},
				"buildingName":    {"b"},
				"streetName":      {"c"},
				"town":            {"d"},
				"postalCode":      {"e"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
					Donor: donordata.Donor{
						Address: place.Address{
							Line1:    "a, b",
							Line2:    "c",
							Line3:    "d",
							Postcode: "e",
							Country:  "FR",
						},
						InternationalAddress: place.InternationalAddress{
							ApartmentNumber: "a",
							BuildingName:    "b",
							StreetName:      "c",
							Town:            "d",
							PostalCode:      "e",
							Country:         "FR",
						},
					},
				}).
				Return(nil)

			err := YourNonUKAddress(nil, donorStore)(tc.appData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					InternationalAddress: place.InternationalAddress{Country: "FR"},
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNonUKAddressWhenLiveInUK(t *testing.T) {
	testCases := map[string]struct {
		url              string
		expectedRedirect string
	}{
		"no from": {
			url:              "/",
			expectedRedirect: donor.PathYourAddress.FormatQuery("lpa-id", url.Values{}),
		},
		"with from": {
			url:              "/?from=/blah",
			expectedRedirect: donor.PathYourAddress.FormatQuery("lpa-id", url.Values{"from": {"/blah"}}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"live-in-uk": {""},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID: "lpa-id",
				}).
				Return(nil)

			err := YourNonUKAddress(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					Address: place.Address{
						Line1:    "a, b",
						Line2:    "c",
						Line3:    "d",
						Postcode: "e",
						Country:  "FR",
					},
					InternationalAddress: place.InternationalAddress{
						ApartmentNumber: "a",
						BuildingName:    "b",
						StreetName:      "c",
						Town:            "d",
						PostalCode:      "e",
						Country:         "FR",
					},
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNonUKAddressWhenAddressNotChanged(t *testing.T) {
	testCases := map[string]struct {
		url              string
		expectedRedirect string
	}{
		"making first LPA": {
			url:              "/",
			expectedRedirect: donor.PathReceivingUpdatesAboutYourLpa.Format("lpa-id"),
		},
		"making another LPA": {
			url:              "/?makingAnotherLPA=1",
			expectedRedirect: donor.PathMakeANewLPA.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				"apartmentNumber": {"a"},
				"buildingName":    {"b"},
				"streetName":      {"c"},
				"town":            {"d"},
				"postalCode":      {"e"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, tc.url, strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := YourNonUKAddress(nil, nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					Address: place.Address{
						Line1:    "a, b",
						Line2:    "c",
						Line3:    "d",
						Postcode: "e",
						Country:  "FR",
					},
					InternationalAddress: place.InternationalAddress{
						ApartmentNumber: "a",
						BuildingName:    "b",
						StreetName:      "c",
						Town:            "d",
						PostalCode:      "e",
						Country:         "FR",
					},
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostYourNonUKAddressWhenStoreErrors(t *testing.T) {
	testcases := map[string]url.Values{
		"submit": {
			"buildingName": {"x"},
			"town":         {"y"},
		},
		"live-in-uk": {
			"live-in-uk": {""},
		},
	}

	for name, f := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), mock.Anything).
				Return(expectedError)

			err := YourNonUKAddress(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestPostYourNonUKAddressWhenValidationError(t *testing.T) {
	f := url.Values{
		"buildingName": {"x"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *yourNonUKAddressData) bool {
			return assert.Equal(t, data.Errors, validation.With("town", validation.EnterError{Label: "townSuburbOrCity"}))
		})).
		Return(nil)

	err := YourNonUKAddress(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadYourNonUkAddressForm(t *testing.T) {
	f := url.Values{
		"apartmentNumber": {"a"},
		"buildingNumber":  {"b"},
		"buildingName":    {"c"},
		"streetName":      {"d"},
		"town":            {"e"},
		"region":          {"f"},
		"postalCode":      {"g"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourNonUKAddressForm(r)

	assert.Equal(t, &yourNonUKAddressForm{
		Address: place.InternationalAddress{
			ApartmentNumber: "a",
			BuildingNumber:  "b",
			BuildingName:    "c",
			StreetName:      "d",
			Town:            "e",
			Region:          "f",
			PostalCode:      "g",
		},
	}, result)
}

func TestYourNonUKAddressFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *yourNonUKAddressForm
		errors validation.List
	}{
		"valid with apartment number": {
			form: &yourNonUKAddressForm{
				Address: place.InternationalAddress{
					ApartmentNumber: "123",
					Town:            "a-town",
				},
			},
		},
		"valid with building number": {
			form: &yourNonUKAddressForm{
				Address: place.InternationalAddress{
					BuildingNumber: "123",
					Town:           "a-town",
				},
			},
		},
		"valid with building name": {
			form: &yourNonUKAddressForm{
				Address: place.InternationalAddress{
					BuildingName: "123",
					Town:         "a-town",
				},
			},
		},
		"missing required": {
			form: &yourNonUKAddressForm{},
			errors: validation.With("buildingAddress", validation.EnterError{Label: "atLeastOneBuildingAddress"}).
				With("town", validation.EnterError{Label: "townSuburbOrCity"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
