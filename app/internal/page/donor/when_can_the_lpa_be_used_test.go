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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Lpa: &page.Lpa{},
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
		Return(&page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:  appData,
			When: page.UsedWhenRegistered,
			Lpa:  &page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered},
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
		Return(&page.Lpa{}, expectedError)

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
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App: appData,
			Lpa: &page.Lpa{},
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
		"when": {page.UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			WhenCanTheLpaBeUsed: page.UsedWhenRegistered,
			Tasks:               page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted, WhenCanTheLpaBeUsed: page.TaskCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.Restrictions, resp.Header.Get("Location"))
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
		Return(&page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted, WhenCanTheLpaBeUsed: page.TaskInProgress},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.Restrictions, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"when": {page.UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered, Tasks: page.Tasks{WhenCanTheLpaBeUsed: page.TaskCompleted}}).
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
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &whenCanTheLpaBeUsedData{
			App:    appData,
			Errors: validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}),
			Lpa:    &page.Lpa{},
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
		"when":         {page.UsedWhenRegistered},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, page.UsedWhenRegistered, result.When)
	assert.True(t, result.AnswerLater)
}

func TestWhenCanTheLpaBeUsedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whenCanTheLpaBeUsedForm
		errors validation.List
	}{
		"when-registered": {
			form: &whenCanTheLpaBeUsedForm{
				When: page.UsedWhenRegistered,
			},
		},
		"when-capacity-lost": {
			form: &whenCanTheLpaBeUsedForm{
				When: page.UsedWhenCapacityLost,
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
