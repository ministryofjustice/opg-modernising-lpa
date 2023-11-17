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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App:     testAppData,
			Lpa:     &page.Lpa{},
			Form:    &whenCanTheLpaBeUsedForm{},
			Options: actor.CanBeUsedWhenValues,
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &whenCanTheLpaBeUsedData{
			App: testAppData,
			Lpa: &page.Lpa{WhenCanTheLpaBeUsed: actor.CanBeUsedWhenHasCapacity},
			Form: &whenCanTheLpaBeUsedForm{
				When: actor.CanBeUsedWhenHasCapacity,
			},
			Options: actor.CanBeUsedWhenValues,
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &page.Lpa{WhenCanTheLpaBeUsed: actor.CanBeUsedWhenHasCapacity})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	form := url.Values{
		"when": {actor.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                  "lpa-id",
			WhenCanTheLpaBeUsed: actor.CanBeUsedWhenHasCapacity,
			Tasks:               page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, WhenCanTheLpaBeUsed: actor.TaskCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &page.Lpa{
		ID:    "lpa-id",
		Tasks: page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"when": {actor.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{WhenCanTheLpaBeUsed: actor.CanBeUsedWhenHasCapacity, Tasks: page.Tasks{WhenCanTheLpaBeUsed: actor.TaskCompleted}}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *whenCanTheLpaBeUsedData) bool {
			return assert.Equal(t, validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}), data.Errors)
		})).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when": {actor.CanBeUsedWhenHasCapacity.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, actor.CanBeUsedWhenHasCapacity, result.When)
}

func TestWhenCanTheLpaBeUsedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whenCanTheLpaBeUsedForm
		errors validation.List
	}{
		"valid": {
			form: &whenCanTheLpaBeUsedForm{},
		},
		"error": {
			form: &whenCanTheLpaBeUsedForm{
				Error: expectedError,
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
