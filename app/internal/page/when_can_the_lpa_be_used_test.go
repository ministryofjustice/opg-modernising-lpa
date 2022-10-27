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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{WhenCanTheLpaBeUsed: "when-registered"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:  appData,
			When: "when-registered",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{WhenCanTheLpaBeUsed: "when-registered", Tasks: Tasks{WhenCanTheLpaBeUsed: TaskCompleted}}).
		Return(nil)

	form := url.Values{
		"when": {"when-registered"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, restrictionsPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenAnswerLater(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{Tasks: Tasks{WhenCanTheLpaBeUsed: TaskInProgress}}).
		Return(nil)

	form := url.Values{
		"when":         {"what"},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, restrictionsPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{WhenCanTheLpaBeUsed: "when-registered", Tasks: Tasks{WhenCanTheLpaBeUsed: TaskCompleted}}).
		Return(expectedError)

	form := url.Values{
		"when": {"when-registered"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when":         {"when-registered"},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, "when-registered", result.When)
	assert.True(t, result.AnswerLater)
}

func TestWhenCanTheLpaBeUsedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whenCanTheLpaBeUsedForm
		errors map[string]string
	}{
		"when-registered": {
			form: &whenCanTheLpaBeUsedForm{
				When: "when-registered",
			},
			errors: map[string]string{},
		},
		"when-capacity-lost": {
			form: &whenCanTheLpaBeUsedForm{
				When: "when-capacity-lost",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &whenCanTheLpaBeUsedForm{},
			errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		},
		"invalid": {
			form: &whenCanTheLpaBeUsedForm{
				When: "what",
			},
			errors: map[string]string{
				"when": "selectWhenCanTheLpaBeUsed",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
