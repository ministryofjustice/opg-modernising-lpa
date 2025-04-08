package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhatCountryDoYouLiveIn(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whatCountryDoYouLiveInData{
			App:       testAppData,
			Form:      &whatCountryDoYouLiveInForm{},
			Countries: place.Countries,
		}).
		Return(nil)

	err := WhatCountryDoYouLiveIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhatCountryDoYouLiveInWithStoreData(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whatCountryDoYouLiveInData{
			App:       testAppData,
			Form:      &whatCountryDoYouLiveInForm{CountryCode: "FR"},
			Countries: place.Countries,
		}).
		Return(nil)

	err := WhatCountryDoYouLiveIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{
			InternationalAddress: place.InternationalAddress{
				Country: "FR",
			},
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhatCountryDoYouLiveInOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatCountryDoYouLiveIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhatCountryDoYouLiveIn(t *testing.T) {
	testcases := map[string]struct {
		redirect donor.Path
		updated  *donordata.Provided
	}{
		"GB": {
			redirect: donor.PathYourAddress,
			updated: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{},
			},
		},
		"FR": {
			redirect: donor.PathYourNonUKAddress,
			updated: &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					InternationalAddress: place.InternationalAddress{
						Country: "FR",
					},
				},
			},
		},
	}

	for country, tc := range testcases {
		t.Run(country, func(t *testing.T) {
			form := url.Values{
				"country": {country},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), tc.updated).
				Return(nil)

			err := WhatCountryDoYouLiveIn(nil, donorStore)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					Address:              place.Address{Line1: "A"},
					InternationalAddress: place.InternationalAddress{Town: "B"},
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, tc.redirect.Format("lpa-id"), resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostWhatCountryDoYouLiveInOnStoreError(t *testing.T) {
	form := url.Values{
		"country": {"FR"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WhatCountryDoYouLiveIn(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhatCountryDoYouLiveInOnInvalidForm(t *testing.T) {
	form := url.Values{
		"country": {"Other"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("country", validation.SelectError{Label: "countryYouLiveIn"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whatCountryDoYouLiveInData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := WhatCountryDoYouLiveIn(template.Execute, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhatCountryYouLiveInForm(t *testing.T) {
	form := url.Values{
		"country": {"DE"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhatCountryDoYouLiveInForm(r)

	assert.Equal(t, result, &whatCountryDoYouLiveInForm{CountryCode: "DE"})
}

func TestWhatCountryDoYouLiveInFormValidate(t *testing.T) {
	testcases := map[string]struct {
		form   *whatCountryDoYouLiveInForm
		errors validation.List
	}{
		"valid": {
			form: &whatCountryDoYouLiveInForm{CountryCode: "GB"},
		},
		"invalid": {
			form:   &whatCountryDoYouLiveInForm{CountryCode: "What"},
			errors: validation.With("country", validation.SelectError{Label: "countryYouLiveIn"}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
