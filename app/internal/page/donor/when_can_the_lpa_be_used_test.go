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
)

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App:  testAppData,
			When: page.UsedWhenRegistered,
			Lpa:  &page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	form := url.Values{
		"when": {page.UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Tasks: page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
		}, nil)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			WhenCanTheLpaBeUsed: page.UsedWhenRegistered,
			Tasks:               page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, WhenCanTheLpaBeUsed: actor.TaskCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"when": {page.UsedWhenRegistered},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	donorStore.
		On("Put", r.Context(), &page.Lpa{WhenCanTheLpaBeUsed: page.UsedWhenRegistered, Tasks: page.Tasks{WhenCanTheLpaBeUsed: actor.TaskCompleted}}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App:    testAppData,
			Errors: validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}),
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when": {page.UsedWhenRegistered},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, page.UsedWhenRegistered, result.When)
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
