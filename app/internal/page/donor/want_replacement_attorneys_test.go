package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWantReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWantReplacementAttorneysWithExistingReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{
				{FirstNames: "this"},
			},
		}, nil)

	template := &mockTemplate{}

	err := WantReplacementAttorneys(template.Func, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWantReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{WantReplacementAttorneys: "yes"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App:  testAppData,
			Want: "yes",
			Lpa:  &page.Lpa{WantReplacementAttorneys: "yes"},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := WantReplacementAttorneys(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := WantReplacementAttorneys(template.Func, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostWantReplacementAttorneys(t *testing.T) {
	testCases := []struct {
		Want                         string
		ExpectedRedirect             string
		ExistingReplacementAttorneys actor.Attorneys
		ExpectedReplacementAttorneys actor.Attorneys
		TaskState                    page.TaskState
	}{
		{
			Want:                         "yes",
			ExpectedRedirect:             "/lpa/lpa-id" + page.Paths.ChooseReplacementAttorneys,
			ExistingReplacementAttorneys: actor.Attorneys{{ID: "123"}},
			ExpectedReplacementAttorneys: actor.Attorneys{{ID: "123"}},
			TaskState:                    page.TaskInProgress,
		},
		{
			Want:             "no",
			ExpectedRedirect: "/lpa/lpa-id" + page.Paths.TaskList,
			ExistingReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "345"},
			},
			ExpectedReplacementAttorneys: actor.Attorneys{},
			TaskState:                    page.TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Want, func(t *testing.T) {
			form := url.Values{
				"want": {tc.Want},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys: tc.ExistingReplacementAttorneys,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &page.Lpa{
					WantReplacementAttorneys: tc.Want,
					ReplacementAttorneys:     tc.ExpectedReplacementAttorneys,
					Tasks:                    page.Tasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			err := WantReplacementAttorneys(nil, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"want": {"yes"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WantReplacementAttorneys(nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostWantReplacementAttorneysWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App:    testAppData,
			Errors: validation.With("want", validation.SelectError{Label: "yesToAddReplacementAttorneys"}),
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWantReplacementAttorneysForm(t *testing.T) {
	form := url.Values{
		"want": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWantReplacementAttorneysForm(r)

	assert.Equal(t, "yes", result.Want)
}

func TestWantReplacementAttorneysFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *wantReplacementAttorneysForm
		errors validation.List
	}{
		"yes": {
			form: &wantReplacementAttorneysForm{
				Want: "yes",
			},
		},
		"no": {
			form: &wantReplacementAttorneysForm{
				Want: "no",
			},
		},
		"missing": {
			form:   &wantReplacementAttorneysForm{},
			errors: validation.With("want", validation.SelectError{Label: "yesToAddReplacementAttorneys"}),
		},
		"invalid": {
			form: &wantReplacementAttorneysForm{
				Want: "what",
			},
			errors: validation.With("want", validation.SelectError{Label: "yesToAddReplacementAttorneys"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
