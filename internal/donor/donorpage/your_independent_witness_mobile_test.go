package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourIndependentWitnessMobile(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourIndependentWitnessMobileData{
			App:  testAppData,
			Form: &yourIndependentWitnessMobileForm{},
		}).
		Return(nil)

	err := YourIndependentWitnessMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourIndependentWitnessMobileFromStore(t *testing.T) {
	testcases := map[string]struct {
		donor *donordata.Provided
		form  *yourIndependentWitnessMobileForm
	}{
		"uk mobile": {
			donor: &donordata.Provided{
				IndependentWitness: donordata.IndependentWitness{
					Mobile: "07777",
				},
			},
			form: &yourIndependentWitnessMobileForm{
				Mobile: "07777",
			},
		},
		"non-uk mobile": {
			donor: &donordata.Provided{
				IndependentWitness: donordata.IndependentWitness{
					Mobile:         "07777",
					HasNonUKMobile: true,
				},
			},
			form: &yourIndependentWitnessMobileForm{
				NonUKMobile:    "07777",
				HasNonUKMobile: true,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &yourIndependentWitnessMobileData{
					App:  testAppData,
					Form: tc.form,
				}).
				Return(nil)

			err := YourIndependentWitnessMobile(template.Execute, nil)(testAppData, w, r, tc.donor)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetYourIndependentWitnessMobileWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &yourIndependentWitnessMobileData{
			App:  testAppData,
			Form: &yourIndependentWitnessMobileForm{},
		}).
		Return(expectedError)

	err := YourIndependentWitnessMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourIndependentWitnessMobile(t *testing.T) {
	testCases := map[string]struct {
		form                         url.Values
		yourIndependentWitnessMobile donordata.IndependentWitness
	}{
		"valid": {
			form: url.Values{
				"mobile": {"07535111111"},
			},
			yourIndependentWitnessMobile: donordata.IndependentWitness{
				Mobile: "07535111111",
			},
		},
		"valid non uk mobile": {
			form: url.Values{
				"has-non-uk-mobile": {"1"},
				"non-uk-mobile":     {"+337575757"},
			},
			yourIndependentWitnessMobile: donordata.IndependentWitness{
				Mobile:         "+337575757",
				HasNonUKMobile: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), &donordata.Provided{
					LpaID:              "lpa-id",
					IndependentWitness: tc.yourIndependentWitnessMobile,
				}).
				Return(nil)

			err := YourIndependentWitnessMobile(nil, donorStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YourIndependentWitnessAddress.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostYourIndependentWitnessMobileWhenValidationError(t *testing.T) {
	form := url.Values{
		"mobile": {"xyz"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *yourIndependentWitnessMobileData) bool {
			return assert.Equal(t, validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}), data.Errors)
		})).
		Return(nil)

	err := YourIndependentWitnessMobile(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourIndependentWitnessMobileWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := YourIndependentWitnessMobile(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestReadYourIndependentWitnessMobileForm(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readYourIndependentWitnessMobileForm(r)

	assert.Equal(t, "07535111111", result.Mobile)
}

func TestYourIndependentWitnessMobileFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *yourIndependentWitnessMobileForm
		errors validation.List
	}{
		"valid": {
			form: &yourIndependentWitnessMobileForm{
				Mobile: "07535111111",
			},
		},
		"missing all": {
			form: &yourIndependentWitnessMobileForm{},
			errors: validation.
				With("mobile", validation.EnterError{Label: "aUKMobileNumber"}),
		},
		"missing when non uk mobile": {
			form: &yourIndependentWitnessMobileForm{HasNonUKMobile: true},
			errors: validation.
				With("non-uk-mobile", validation.EnterError{Label: "aMobilePhoneNumber"}),
		},
		"invalid incorrect mobile format": {
			form: &yourIndependentWitnessMobileForm{
				Mobile: "0753511111",
			},
			errors: validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
		"invalid non uk mobile format": {
			form: &yourIndependentWitnessMobileForm{
				HasNonUKMobile: true,
				NonUKMobile:    "0753511111",
			},
			errors: validation.With("non-uk-mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
