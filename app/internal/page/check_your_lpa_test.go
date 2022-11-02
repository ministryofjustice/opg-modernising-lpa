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

func TestGetCheckYourLpa(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &checkYourLpaData{
			App:                           appData,
			Form:                          &checkYourLpaForm{},
			Lpa:                           &Lpa{},
			HowAttorneysMakeDecisionsPath: howShouldAttorneysMakeDecisionsPath,
			ChooseAttorneysPath:           chooseAttorneysPath,
			WhenCanLpaBeUsedPath:          whenCanTheLpaBeUsedPath,
			RestrictionsPath:              restrictionsPath,
			CertificatesProviderPath:      certificateProviderDetailsPath,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CheckYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetCheckYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CheckYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCheckYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{
		Checked:      true,
		HappyToShare: true,
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &checkYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				Checked: true,
				Happy:   true,
			},
			HowAttorneysMakeDecisionsPath: howShouldAttorneysMakeDecisionsPath,
			ChooseAttorneysPath:           chooseAttorneysPath,
			WhenCanLpaBeUsedPath:          whenCanTheLpaBeUsedPath,
			RestrictionsPath:              restrictionsPath,
			CertificatesProviderPath:      certificateProviderDetailsPath,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := CheckYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostCheckYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := &Lpa{
		Checked:      false,
		HappyToShare: false,
		Tasks:        Tasks{CheckYourLpa: TaskInProgress},
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			Checked:      true,
			HappyToShare: true,
			Tasks:        Tasks{CheckYourLpa: TaskCompleted},
		}).
		Return(nil)

	form := url.Values{
		"checked": {"1"},
		"happy":   {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CheckYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whatHappensNextPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCheckYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{
			Checked:      true,
			HappyToShare: true,
			Tasks:        Tasks{CheckYourLpa: TaskCompleted},
		}).
		Return(expectedError)

	form := url.Values{
		"checked": {"1"},
		"happy":   {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CheckYourLpa(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostCheckYourLpaWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *checkYourLpaData) bool {
			return assert.Equal(t, map[string]string{"happy": "selectHappyToShareLpa"}, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"checked": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := CheckYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadCheckYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"checked": {" 1   "},
		"happy":   {" 0"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readCheckYourLpaForm(r)

	assert.Equal(true, result.Checked)
	assert.Equal(false, result.Happy)
}

func TestCheckYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *checkYourLpaForm
		errors map[string]string
	}{
		"valid": {
			form: &checkYourLpaForm{
				Happy:   true,
				Checked: true,
			},
			errors: map[string]string{},
		},
		"invalid-all": {
			form: &checkYourLpaForm{
				Happy:   false,
				Checked: false,
			},
			errors: map[string]string{
				"happy":   "selectHappyToShareLpa",
				"checked": "selectCheckedLpa",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
