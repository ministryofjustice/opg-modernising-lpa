package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSignYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &signYourLpaData{
			App:                  page.TestAppData,
			Form:                 &signYourLpaForm{},
			Lpa:                  &page.Lpa{},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetSignYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := SignYourLpa(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetSignYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		WantToSignLpa:     true,
		WantToApplyForLpa: false,
	}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &signYourLpaData{
			App: page.TestAppData,
			Lpa: lpa,
			Form: &signYourLpaForm{
				WantToSign:  true,
				WantToApply: false,
			},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostSignYourLpa(t *testing.T) {
	f := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			WantToSignLpa:     true,
			WantToApplyForLpa: true,
		}).
		Return(nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{
				ConfirmYourIdentityAndSign: page.TaskCompleted,
			},
			WantToSignLpa:     true,
			WantToApplyForLpa: true,
		}).
		Return(nil)

	err := SignYourLpa(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.WitnessingYourSignature, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(page.ExpectedError)

	err := SignYourLpa(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostSignYourLpaWhenValidationErrors(t *testing.T) {
	f := url.Values{
		"sign-lpa": {"unrecognised-signature", "another-unrecognised-signature"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			WantToSignLpa:     false,
			WantToApplyForLpa: false,
		}).
		Return(nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *signYourLpaData) bool {
			return assert.Equal(t, validation.With("sign-lpa", validation.CustomError{Label: "bothBoxesToSignAndApply"}), data.Errors)
		})).
		Return(nil)

	err := SignYourLpa(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadSignYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	f := url.Values{
		"sign-lpa": {"want-to-sign", "want-to-apply"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readSignYourLpaForm(r)

	assert.Equal(true, result.WantToSign)
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
				WantToSign:  true,
			},
		},
		"only want-to-sign selected": {
			form: &signYourLpaForm{
				WantToApply: false,
				WantToSign:  true,
			},
			errors: validation.With("sign-lpa", validation.CustomError{Label: "bothBoxesToSignAndApply"}),
		},
		"only want-to-apply selected": {
			form: &signYourLpaForm{
				WantToApply: true,
				WantToSign:  false,
			},
			errors: validation.With("sign-lpa", validation.CustomError{Label: "bothBoxesToSignAndApply"}),
		},
		"none selected": {
			form:   &signYourLpaForm{},
			errors: validation.With("sign-lpa", validation.CustomError{Label: "bothBoxesToSignAndApply"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
