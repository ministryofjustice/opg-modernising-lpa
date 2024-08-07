package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWitnessingAsIndependentWitness(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  &witnessingAsIndependentWitnessForm{},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, nil, time.Now)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsIndependentWitnessFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App: testAppData,
			Donor: &donordata.Provided{
				IndependentWitness: donordata.IndependentWitness{FirstNames: "Joan"},
			},
			Form: &witnessingAsIndependentWitnessForm{},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, nil, time.Now)(testAppData, w, r, &donordata.Provided{
		IndependentWitness: donordata.IndependentWitness{FirstNames: "Joan"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWitnessingAsIndependentWitnessWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  &witnessingAsIndependentWitnessForm{},
		}).
		Return(expectedError)

	err := WitnessingAsIndependentWitness(template.Execute, nil, time.Now)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsIndependentWitness(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:                           "lpa-id",
			DonorIdentityUserData:           identity.UserData{Status: identity.StatusConfirmed},
			IndependentWitnessCodes:         donordata.WitnessCodes{{Code: "1234", Created: now}},
			WitnessedByIndependentWitnessAt: now,
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(nil, donorStore, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{
		LpaID:                   "lpa-id",
		DonorIdentityUserData:   identity.UserData{Status: identity.StatusConfirmed},
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathWitnessingAsCertificateProvider.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWitnessingAsIndependentWitnessWhenDonorStoreErrors(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)
	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := WitnessingAsIndependentWitness(nil, donorStore, func() time.Time { return now })(testAppData, w, r, &donordata.Provided{
		LpaID:                   "lpa-id",
		DonorIdentityUserData:   identity.UserData{Status: identity.StatusConfirmed},
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
	})
	assert.Equal(t, expectedError, err)
}

func TestPostWitnessingAsIndependentWitnessCodeTooOld(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App: testAppData,
			Donor: &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsIndependentWitnessForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, donorStore, time.Now)(testAppData, w, r, &donordata.Provided{
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsIndependentWitnessCodeDoesNotMatch(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App: testAppData,
			Donor: &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeDoesNotMatch"}),
			Form:   &witnessingAsIndependentWitnessForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, donorStore, time.Now)(testAppData, w, r, &donordata.Provided{
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsIndependentWitnessWhenCodeExpired(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()
	invalidCreated := now.Add(-45 * time.Minute)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App: testAppData,
			Donor: &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "witnessCodeExpired"}),
			Form:   &witnessingAsIndependentWitnessForm{Code: "1234"},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, donorStore, time.Now)(testAppData, w, r, &donordata.Provided{
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: invalidCreated}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWitnessingAsIndependentWitnessCodeLimitBreached(t *testing.T) {
	form := url.Values{
		"witness-code": {"4321"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	now := time.Now()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.MatchedBy(func(donor *donordata.Provided) bool {
			donor.WitnessCodeLimiter = nil
			return assert.Equal(t, donor, &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			})
		})).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &witnessingAsIndependentWitnessData{
			App: testAppData,
			Donor: &donordata.Provided{
				IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
			},
			Errors: validation.With("witness-code", validation.CustomError{Label: "tooManyWitnessCodeAttempts"}),
			Form:   &witnessingAsIndependentWitnessForm{Code: "4321"},
		}).
		Return(nil)

	err := WitnessingAsIndependentWitness(template.Execute, donorStore, time.Now)(testAppData, w, r, &donordata.Provided{
		WitnessCodeLimiter:      donordata.NewLimiter(time.Minute, 0, 10),
		IndependentWitnessCodes: donordata.WitnessCodes{{Code: "1234", Created: now}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWitnessingAsIndependentWitnessForm(t *testing.T) {
	form := url.Values{
		"witness-code": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWitnessingAsIndependentWitnessForm(r)

	assert.Equal(t, "1234", result.Code)
}

func TestWitnessingAsIndependentWitnessValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *witnessingAsIndependentWitnessForm
		errors validation.List
	}{
		"valid numeric": {
			form: &witnessingAsIndependentWitnessForm{
				Code: "1234",
			},
		},
		"valid alpha": {
			form: &witnessingAsIndependentWitnessForm{
				Code: "aBcD",
			},
		},
		"missing": {
			form:   &witnessingAsIndependentWitnessForm{},
			errors: validation.With("witness-code", validation.EnterError{Label: "theCodeWeSentIndependentWitness"}),
		},
		"too long": {
			form: &witnessingAsIndependentWitnessForm{
				Code: "12345",
			},
			errors: validation.With("witness-code", validation.StringLengthError{Label: "theCodeWeSentIndependentWitness", Length: 4}),
		},
		"too short": {
			form: &witnessingAsIndependentWitnessForm{
				Code: "123",
			},
			errors: validation.With("witness-code", validation.StringLengthError{Label: "theCodeWeSentIndependentWitness", Length: 4}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
