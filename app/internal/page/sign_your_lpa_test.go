package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
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

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SignYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetSignYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetSignYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{
		CPWitnessedDonorSign: true,
		WantToApplyForLpa:    false,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
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

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := SignYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			CPWitnessedDonorSign: true,
			WantToApplyForLpa:    true,
		}).
		Return(nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			Tasks: Tasks{
				ConfirmYourIdentityAndSign: TaskCompleted,
			},
			CPWitnessedDonorSign: true,
			WantToApplyForLpa:    true,
		}).
		Return(nil)

	form := url.Values{
		"sign-lpa": {"cp-witnessed", "want-to-apply"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.WitnessingYourSignature, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"sign-lpa": {"cp-witnessed", "want-to-apply"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := SignYourLpa(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			CPWitnessedDonorSign: false,
			WantToApplyForLpa:    false,
		}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *signYourLpaData) bool {
			return assert.Equal(t, map[string]string{"sign-lpa": "selectBothBoxes"}, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"sign-lpa": {"unrecognised-signature", "another-unrecognised-signature"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

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
		errors map[string]string
	}{
		"valid": {
			form: &signYourLpaForm{
				WantToApply: true,
				CPWitnessed: true,
			},
			errors: map[string]string{},
		},
		"only cp-witnessed selected": {
			form: &signYourLpaForm{
				WantToApply: false,
				CPWitnessed: true,
			},
			errors: map[string]string{
				"sign-lpa": "selectBothBoxes",
			},
		},
		"only want-to-apply selected": {
			form: &signYourLpaForm{
				WantToApply: true,
				CPWitnessed: false,
			},
			errors: map[string]string{
				"sign-lpa": "selectBothBoxes",
			},
		},
		"none selected": {
			form: &signYourLpaForm{},
			errors: map[string]string{
				"sign-lpa": "selectBothBoxes",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
