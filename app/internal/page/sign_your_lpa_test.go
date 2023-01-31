package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &signYourLpaData{
			App:                  appData,
			Form:                 &signYourLpaForm{},
			Lpa:                  &Lpa{},
			CPWitnessedFormValue: CertificateProviderHasWitnessed,
			WantFormValue:        WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetSignYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetSignYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{
		CPWitnessedDonorSign: true,
		WantToApplyForLpa:    false,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &signYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &signYourLpaForm{
				CPWitnessed: true,
				WantToApply: false,
			},
			CPWitnessedFormValue: CertificateProviderHasWitnessed,
			WantFormValue:        WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostSignYourLpa(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"cp-witnessed", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			CPWitnessedDonorSign: true,
			WantToApplyForLpa:    true,
		}).
		Return(nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			Tasks: Tasks{
				ConfirmYourIdentityAndSign: TaskCompleted,
			},
			CPWitnessedDonorSign: true,
			WantToApplyForLpa:    true,
		}).
		Return(nil)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.WitnessingYourSignature, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"cp-witnessed", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"sign-lpa": {"unrecognised-signature", "another-unrecognised-signature"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			CPWitnessedDonorSign: false,
			WantToApplyForLpa:    false,
		}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *signYourLpaData) bool {
			return assert.Equal(t, validation.With("sign-lpa", validation.SelectedError{Label: "bothBoxesToSign"}), data.Errors)
		})).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadSignYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"sign-lpa": {"cp-witnessed", "want-to-apply"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readSignYourLpaForm(r)

	assert.Equal(true, result.CPWitnessed)
	assert.Equal(true, result.WantToApply)
}

func TestSignYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *signYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &signYourLpaForm{
				WantToApply: true,
				CPWitnessed: true,
			},
		},
		"only cp-witnessed selected": {
			form: &signYourLpaForm{
				WantToApply: false,
				CPWitnessed: true,
			},
			errors: validation.With("sign-lpa", validation.SelectedError{Label: "bothBoxesToSign"}),
		},
		"only want-to-apply selected": {
			form: &signYourLpaForm{
				WantToApply: true,
				CPWitnessed: false,
			},
			errors: validation.With("sign-lpa", validation.SelectedError{Label: "bothBoxesToSign"}),
		},
		"none selected": {
			form:   &signYourLpaForm{},
			errors: validation.With("sign-lpa", validation.SelectedError{Label: "bothBoxesToSign"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
