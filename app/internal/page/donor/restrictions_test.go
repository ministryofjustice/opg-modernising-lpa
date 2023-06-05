package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetRestrictions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &restrictionsData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(nil)

	err := Restrictions(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRestrictionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Restrictions: "blah"}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &restrictionsData{
			App: testAppData,
			Lpa: &page.Lpa{Restrictions: "blah"},
		}).
		Return(nil)

	err := Restrictions(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRestrictionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, expectedError)

	err := Restrictions(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRestrictionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &restrictionsData{
			App: testAppData,
			Lpa: &page.Lpa{},
		}).
		Return(expectedError)

	err := Restrictions(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostRestrictions(t *testing.T) {
	form := url.Values{
		"restrictions": {"blah"},
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
			Restrictions: "blah",
			Tasks:        page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, Restrictions: actor.TaskCompleted},
		}).
		Return(nil)

	err := Restrictions(nil, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostRestrictionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"restrictions": {"blah"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	donorStore.
		On("Put", r.Context(), &page.Lpa{Restrictions: "blah", Tasks: page.Tasks{Restrictions: actor.TaskCompleted}}).
		Return(expectedError)

	err := Restrictions(nil, donorStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostRestrictionsWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"restrictions": {random.String(10001)},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &restrictionsData{
			App:    testAppData,
			Errors: validation.With("restrictions", validation.StringTooLongError{Label: "restrictions", Length: 10000}),
			Lpa:    &page.Lpa{},
		}).
		Return(nil)

	err := Restrictions(template.Execute, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadRestrictionsForm(t *testing.T) {
	form := url.Values{
		"restrictions": {"blah"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readRestrictionsForm(r)

	assert.Equal(t, "blah", result.Restrictions)
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
