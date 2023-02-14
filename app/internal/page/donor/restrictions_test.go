package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRestrictions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App: page.TestAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := Restrictions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetRestrictionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Restrictions: "blah"}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App: page.TestAppData,
			Lpa: &page.Lpa{Restrictions: "blah"},
		}).
		Return(nil)

	err := Restrictions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetRestrictionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := Restrictions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetRestrictionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App: page.TestAppData,
			Lpa: &page.Lpa{},
		}).
		Return(page.ExpectedError)

	err := Restrictions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostRestrictions(t *testing.T) {
	f := url.Values{
		"restrictions": {"blah"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Restrictions: "blah",
			Tasks:        page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted, Restrictions: page.TaskCompleted},
		}).
		Return(nil)

	err := Restrictions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.WhoDoYouWantToBeCertificateProviderGuidance, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenAnswerLater(t *testing.T) {
	f := url.Values{
		"restrictions": {"what"},
		"answer-later": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Tasks: page.Tasks{YourDetails: page.TaskCompleted, ChooseAttorneys: page.TaskCompleted, Restrictions: page.TaskInProgress},
		}).
		Return(nil)

	err := Restrictions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.WhoDoYouWantToBeCertificateProviderGuidance, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"restrictions": {"blah"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{Restrictions: "blah", Tasks: page.Tasks{Restrictions: page.TaskCompleted}}).
		Return(page.ExpectedError)

	err := Restrictions(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostRestrictionsWhenValidationErrors(t *testing.T) {
	f := url.Values{
		"restrictions": {random.String(10001)},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &restrictionsData{
			App:    page.TestAppData,
			Errors: validation.With("restrictions", validation.StringTooLongError{Label: "restrictions", Length: 10000}),
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := Restrictions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadRestrictionsForm(t *testing.T) {
	f := url.Values{
		"restrictions": {"blah"},
		"answer-later": {"1"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readRestrictionsForm(r)

	assert.Equal(t, "blah", result.Restrictions)
	assert.True(t, result.AnswerLater)
}

func TestRestrictionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *restrictionsForm
		errors validation.List
	}{
		"set": {
			form: &restrictionsForm{
				Restrictions: "blah",
			},
		},
		"too-long": {
			form: &restrictionsForm{
				Restrictions: random.String(10001),
			},
			errors: validation.With("restrictions", validation.StringTooLongError{Label: "restrictions", Length: 10000}),
		},
		"missing": {
			form: &restrictionsForm{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
