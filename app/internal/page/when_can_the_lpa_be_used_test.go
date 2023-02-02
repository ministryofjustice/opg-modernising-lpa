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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{WhenCanTheLpaBeUsed: UsedWhenRegistered}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:  appData,
			When: UsedWhenRegistered,
			Lpa:  &Lpa{WhenCanTheLpaBeUsed: UsedWhenRegistered},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	form := url.Values{
		"when": {UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Tasks: Tasks{ChooseAttorneys: TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			WhenCanTheLpaBeUsed: UsedWhenRegistered,
			Tasks:               Tasks{ChooseAttorneys: TaskCompleted, WhenCanTheLpaBeUsed: TaskCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.Restrictions, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenAnswerLater(t *testing.T) {
	form := url.Values{
		"when":         {"what"},
		"answer-later": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			Tasks: Tasks{ChooseAttorneys: TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{
			Tasks: Tasks{ChooseAttorneys: TaskCompleted, WhenCanTheLpaBeUsed: TaskInProgress},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.Restrictions, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"when": {UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{WhenCanTheLpaBeUsed: UsedWhenRegistered, Tasks: Tasks{WhenCanTheLpaBeUsed: TaskCompleted}}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:    appData,
			Errors: validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}),
			Lpa:    &Lpa{},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when":         {UsedWhenRegistered},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, UsedWhenRegistered, result.When)
	assert.True(t, result.AnswerLater)
}

func TestWhenCanTheLpaBeUsedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whenCanTheLpaBeUsedForm
		errors validation.List
	}{
		"when-registered": {
			form: &whenCanTheLpaBeUsedForm{
				When: UsedWhenRegistered,
			},
		},
		"when-capacity-lost": {
			form: &whenCanTheLpaBeUsedForm{
				When: UsedWhenCapacityLost,
			},
		},
		"missing": {
			form:   &whenCanTheLpaBeUsedForm{},
			errors: validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}),
		},
		"invalid": {
			form: &whenCanTheLpaBeUsedForm{
				When: "what",
			},
			errors: validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
