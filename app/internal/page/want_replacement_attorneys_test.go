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

func TestGetWantReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysWithExistingReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{
			ReplacementAttorneys: []Attorney{
				{FirstNames: "this"},
			},
		}, nil)

	template := &mockTemplate{}

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{WantReplacementAttorneys: "yes"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App:  appData,
			Want: "yes",
			Lpa:  &Lpa{WantReplacementAttorneys: "yes"},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := WantReplacementAttorneys(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWantReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostWantReplacementAttorneys(t *testing.T) {
	testCases := []struct {
		Want                         string
		ExpectedRedirect             string
		ExistingReplacementAttorneys []Attorney
		ExpectedReplacementAttorneys []Attorney
		TaskState                    TaskState
	}{
		{
			Want:                         "yes",
			ExpectedRedirect:             "/lpa/lpa-id" + Paths.ChooseReplacementAttorneys,
			ExistingReplacementAttorneys: []Attorney{{ID: "123"}},
			ExpectedReplacementAttorneys: []Attorney{{ID: "123"}},
			TaskState:                    TaskInProgress,
		},
		{
			Want:             "no",
			ExpectedRedirect: "/lpa/lpa-id" + Paths.TaskList,
			ExistingReplacementAttorneys: []Attorney{
				{ID: "123"},
				{ID: "345"},
			},
			ExpectedReplacementAttorneys: []Attorney{},
			TaskState:                    TaskCompleted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Want, func(t *testing.T) {
			form := url.Values{
				"want": {tc.Want},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&Lpa{
					ReplacementAttorneys: tc.ExistingReplacementAttorneys,
				}, nil)
			lpaStore.
				On("Put", r.Context(), &Lpa{
					WantReplacementAttorneys: tc.Want,
					ReplacementAttorneys:     tc.ExpectedReplacementAttorneys,
					Tasks:                    Tasks{ChooseReplacementAttorneys: tc.TaskState},
				}).
				Return(nil)

			err := WantReplacementAttorneys(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})

	}
}

func TestPostWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"want": {"yes"},
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

	err := WantReplacementAttorneys(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWantReplacementAttorneysWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App:    appData,
			Errors: validation.With("want", "selectWantReplacementAttorneys"),
			Lpa:    &Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
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
	r.Header.Add("Content-Type", formUrlEncoded)

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
			errors: validation.With("want", "selectWantReplacementAttorneys"),
		},
		"invalid": {
			form: &wantReplacementAttorneysForm{
				Want: "what",
			},
			errors: validation.With("want", "selectWantReplacementAttorneys"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
