package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetLifeSustainingTreatment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lifeSustainingTreatmentData{
			App: testAppData,
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLifeSustainingTreatmentFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lifeSustainingTreatmentData{
			App:    testAppData,
			Option: page.OptionA,
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &page.Lpa{LifeSustainingTreatmentOption: page.OptionA})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLifeSustainingTreatmentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lifeSustainingTreatmentData{
			App: testAppData,
		}).
		Return(expectedError)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLifeSustainingTreatment(t *testing.T) {
	form := url.Values{
		"option": {page.OptionA},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			LifeSustainingTreatmentOption: page.OptionA,
			Tasks:                         page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, LifeSustainingTreatment: actor.TaskCompleted},
		}).
		Return(nil)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &page.Lpa{
		Tasks: page.Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostLifeSustainingTreatmentWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"option": {page.OptionA},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{LifeSustainingTreatmentOption: page.OptionA, Tasks: page.Tasks{LifeSustainingTreatment: actor.TaskCompleted}}).
		Return(expectedError)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostLifeSustainingTreatmentWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lifeSustainingTreatmentData{
			App:    testAppData,
			Errors: validation.With("option", validation.SelectError{Label: "ifTheDonorGivesConsentToLifeSustainingTreatment"}),
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadLifeSustainingTreatmentForm(t *testing.T) {
	form := url.Values{
		"option": {page.OptionA},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readLifeSustainingTreatmentForm(r)

	assert.Equal(t, page.OptionA, result.Option)
}

func TestLifeSustainingTreatmentFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lifeSustainingTreatmentForm
		errors validation.List
	}{
		"option a": {
			form: &lifeSustainingTreatmentForm{
				Option: page.OptionA,
			},
		},
		"option b": {
			form: &lifeSustainingTreatmentForm{
				Option: page.OptionB,
			},
		},
		"missing": {
			form:   &lifeSustainingTreatmentForm{},
			errors: validation.With("option", validation.SelectError{Label: "ifTheDonorGivesConsentToLifeSustainingTreatment"}),
		},
		"invalid": {
			form: &lifeSustainingTreatmentForm{
				Option: "what",
			},
			errors: validation.With("option", validation.SelectError{Label: "ifTheDonorGivesConsentToLifeSustainingTreatment"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
