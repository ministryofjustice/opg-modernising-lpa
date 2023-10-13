package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChangeMobileNumber(t *testing.T) {
	for _, actorType := range []actor.Type{actor.TypeIndependentWitness, actor.TypeCertificateProvider} {
		t.Run(actorType.String(), func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &changeMobileNumberData{
					App:       testAppData,
					Form:      &changeMobileNumberForm{},
					ActorType: actorType,
				}).
				Return(nil)

			err := ChangeMobileNumber(template.Execute, nil, actorType)(testAppData, w, r, &page.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChangeMobileNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		lpa  *page.Lpa
		form *changeMobileNumberForm
	}{
		"uk mobile": {
			lpa: &page.Lpa{
				IndependentWitness: actor.IndependentWitness{
					Mobile: "07777",
				},
			},
			form: &changeMobileNumberForm{
				Mobile: "07777",
			},
		},
		"non-uk mobile": {
			lpa: &page.Lpa{
				IndependentWitness: actor.IndependentWitness{
					Mobile:         "07777",
					HasNonUKMobile: true,
				},
			},
			form: &changeMobileNumberForm{
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
			template.
				On("Execute", w, &changeMobileNumberData{
					App:  testAppData,
					Form: tc.form,
				}).
				Return(nil)

			err := ChangeMobileNumber(template.Execute, nil)(testAppData, w, r, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetChangeMobileNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &changeMobileNumberData{
			App:  testAppData,
			Form: &changeMobileNumberForm{},
		}).
		Return(expectedError)

	err := ChangeMobileNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChangeMobileNumber(t *testing.T) {
	testCases := map[string]struct {
		form               url.Values
		changeMobileNumber actor.IndependentWitness
	}{
		"valid": {
			form: url.Values{
				"mobile": {"07535111111"},
			},
			changeMobileNumber: actor.IndependentWitness{
				Mobile: "07535111111",
			},
		},
		"valid non uk mobile": {
			form: url.Values{
				"has-non-uk-mobile": {"1"},
				"non-uk-mobile":     {"+337575757"},
			},
			changeMobileNumber: actor.IndependentWitness{
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
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					ID:                 "lpa-id",
					IndependentWitness: tc.changeMobileNumber,
				}).
				Return(nil)

			err := ChangeMobileNumber(nil, donorStore)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.YourIndependentWitnessAddress.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostChangeMobileNumberWhenValidationError(t *testing.T) {
	form := url.Values{
		"mobile": {"xyz"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *changeMobileNumberData) bool {
			return assert.Equal(t, validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}), data.Errors)
		})).
		Return(nil)

	err := ChangeMobileNumber(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostChangeMobileNumberWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ChangeMobileNumber(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestReadChangeMobileNumberForm(t *testing.T) {
	form := url.Values{
		"mobile": {"07535111111"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readChangeMobileNumberForm(r)

	assert.Equal(t, "07535111111", result.Mobile)
}

func TestChangeMobileNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *changeMobileNumberForm
		errors validation.List
	}{
		"valid": {
			form: &changeMobileNumberForm{
				Mobile: "07535111111",
			},
		},
		"missing all": {
			form: &changeMobileNumberForm{},
			errors: validation.
				With("mobile", validation.EnterError{Label: "aUKMobileNumber"}),
		},
		"missing when non uk mobile": {
			form: &changeMobileNumberForm{HasNonUKMobile: true},
			errors: validation.
				With("non-uk-mobile", validation.EnterError{Label: "aMobilePhoneNumber"}),
		},
		"invalid incorrect mobile format": {
			form: &changeMobileNumberForm{
				Mobile: "0753511111",
			},
			errors: validation.With("mobile", validation.CustomError{Label: "enterAMobileNumberInTheCorrectFormat"}),
		},
		"invalid non uk mobile format": {
			form: &changeMobileNumberForm{
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
