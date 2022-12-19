package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRestrictions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Restrictions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetRestrictionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{Restrictions: "blah"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App:          appData,
			Restrictions: "blah",
			Lpa:          &Lpa{Restrictions: "blah"},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Restrictions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetRestrictionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Restrictions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetRestrictionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := Restrictions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostRestrictions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{Restrictions: "blah", Tasks: Tasks{Restrictions: TaskCompleted}}).
		Return(nil)

	form := url.Values{
		"restrictions": {"blah"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := Restrictions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenAnswerLater(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{Tasks: Tasks{Restrictions: TaskInProgress}}).
		Return(nil)

	form := url.Values{
		"restrictions": {"what"},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := Restrictions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.WhoDoYouWantToBeCertificateProviderGuidance, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{Restrictions: "blah", Tasks: Tasks{Restrictions: TaskCompleted}}).
		Return(expectedError)

	form := url.Values{
		"restrictions": {"blah"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := Restrictions(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	restrictions := random.String(10001)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App:          appData,
			Restrictions: restrictions,
			Errors: map[string]string{
				"restrictions": "restrictionsTooLong",
			},
			Lpa: &Lpa{},
		}).
		Return(nil)

	form := url.Values{
		"restrictions": {restrictions},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := Restrictions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadRestrictionsForm(t *testing.T) {
	form := url.Values{
		"restrictions": {"blah"},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readRestrictionsForm(r)

	assert.Equal(t, "blah", result.Restrictions)
	assert.True(t, result.AnswerLater)
}

func TestRestrictionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *restrictionsForm
		errors map[string]string
	}{
		"set": {
			form: &restrictionsForm{
				Restrictions: "blah",
			},
			errors: map[string]string{},
		},
		"too-long": {
			form: &restrictionsForm{
				Restrictions: random.String(10001),
			},
			errors: map[string]string{
				"restrictions": "restrictionsTooLong",
			},
		},
		"missing": {
			form:   &restrictionsForm{},
			errors: map[string]string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
