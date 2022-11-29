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

func TestGetWantReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysWithExistingReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			ReplacementAttorneys: []Attorney{
				{FirstNames: "this"},
			},
		}, nil)

	template := &mockTemplate{}

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, chooseReplacementAttorneysSummaryPath, resp.Header.Get("Location"))

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{WantReplacementAttorneys: "yes"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App:  appData,
			Want: "yes",
			Lpa:  &Lpa{WantReplacementAttorneys: "yes"},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WantReplacementAttorneys(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := WantReplacementAttorneys(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWantReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: appData,
			Lpa: &Lpa{},
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

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
	}{
		{
			Want:                         "yes",
			ExpectedRedirect:             chooseReplacementAttorneysPath,
			ExistingReplacementAttorneys: []Attorney{{ID: "123"}},
			ExpectedReplacementAttorneys: []Attorney{{ID: "123"}},
		},
		{
			Want:             "no",
			ExpectedRedirect: taskListPath,
			ExistingReplacementAttorneys: []Attorney{
				{ID: "123"},
				{ID: "345"},
			},
			ExpectedReplacementAttorneys: []Attorney{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Want, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					ReplacementAttorneys: tc.ExistingReplacementAttorneys,
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					WantReplacementAttorneys: tc.Want,
					ReplacementAttorneys:     tc.ExpectedReplacementAttorneys,
				}).
				Return(nil)

			form := url.Values{
				"want": {tc.Want},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

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
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{WantReplacementAttorneys: "yes"}).
		Return(expectedError)

	form := url.Values{
		"want": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := WantReplacementAttorneys(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWantReplacementAttorneysWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &wantReplacementAttorneysData{
			App: appData,
			Errors: map[string]string{
				"want": "selectWantReplacementAttorneys",
			},
			Lpa: &Lpa{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

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
		errors map[string]string
	}{
		"yes": {
			form: &wantReplacementAttorneysForm{
				Want: "yes",
			},
			errors: map[string]string{},
		},
		"no": {
			form: &wantReplacementAttorneysForm{
				Want: "no",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &wantReplacementAttorneysForm{},
			errors: map[string]string{
				"want": "selectWantReplacementAttorneys",
			},
		},
		"invalid": {
			form: &wantReplacementAttorneysForm{
				Want: "what",
			},
			errors: map[string]string{
				"want": "selectWantReplacementAttorneys",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
