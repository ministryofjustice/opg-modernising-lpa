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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wantReplacementAttorneysData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWithExistingReplacementAttorneys(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			ReplacementAttorneys: actor.Attorneys{
				{FirstNames: "this"},
			},
		}, nil)

	template := newMockTemplate(t)

	err := WantReplacementAttorneys(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ChooseReplacementAttorneysSummary, resp.Header.Get("Location"))

}

func TestGetWantReplacementAttorneysFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{WantReplacementAttorneys: "yes"}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wantReplacementAttorneysData{
			App:  testAppData,
			Want: "yes",
			Lpa:  &page.Lpa{WantReplacementAttorneys: "yes"},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := WantReplacementAttorneys(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWantReplacementAttorneysWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wantReplacementAttorneysData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := WantReplacementAttorneys(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWantReplacementAttorneys(t *testing.T) {
	testCases := map[string]struct {
		want                         string
		existingReplacementAttorneys actor.Attorneys
		expectedReplacementAttorneys actor.Attorneys
		taskState                    actor.TaskState
		redirectURL                  string
	}{
		"yes": {
			want:                         "yes",
			existingReplacementAttorneys: actor.Attorneys{{ID: "123"}},
			expectedReplacementAttorneys: actor.Attorneys{{ID: "123"}},
			taskState:                    actor.TaskInProgress,
			redirectURL:                  page.Paths.ChooseReplacementAttorneys,
		},
		"no": {
			want: "no",
			existingReplacementAttorneys: actor.Attorneys{
				{ID: "123"},
				{ID: "345"},
			},
			expectedReplacementAttorneys: actor.Attorneys{},
			taskState:                    actor.TaskCompleted,
			redirectURL:                  page.Paths.TaskList,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"want": {tc.want},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Get", r.Context()).
				Return(&page.Lpa{
					ReplacementAttorneys: tc.existingReplacementAttorneys,
					Tasks:                page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
				}, nil)
			donorStore.
				On("Put", r.Context(), &page.Lpa{
					WantReplacementAttorneys: tc.want,
					ReplacementAttorneys:     tc.expectedReplacementAttorneys,
					Tasks:                    page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, ChooseReplacementAttorneys: tc.taskState},
				}).
				Return(nil)

			err := WantReplacementAttorneys(nil, donorStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/lpa/lpa-id"+tc.redirectURL, resp.Header.Get("Location"))
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := WantReplacementAttorneys(nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostWantReplacementAttorneysWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &wantReplacementAttorneysData{
			App:    testAppData,
			Errors: validation.With("want", validation.SelectError{Label: "yesToAddReplacementAttorneys"}),
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := WantReplacementAttorneys(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
