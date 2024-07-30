package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLifeSustainingTreatment(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lifeSustainingTreatmentData{
			App:     testAppData,
			Form:    &lifeSustainingTreatmentForm{},
			Options: donordata.LifeSustainingTreatmentValues,
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLifeSustainingTreatmentFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lifeSustainingTreatmentData{
			App: testAppData,
			Form: &lifeSustainingTreatmentForm{
				Option: actor.LifeSustainingTreatmentOptionA,
			},
			Options: donordata.LifeSustainingTreatmentValues,
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LifeSustainingTreatmentOption: actor.LifeSustainingTreatmentOptionA})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLifeSustainingTreatmentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLifeSustainingTreatment(t *testing.T) {
	form := url.Values{
		"option": {actor.LifeSustainingTreatmentOptionA.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:                         "lpa-id",
			LifeSustainingTreatmentOption: actor.LifeSustainingTreatmentOptionA,
			Tasks:                         actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, LifeSustainingTreatment: actor.TaskCompleted},
		}).
		Return(nil)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostLifeSustainingTreatmentWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"option": {actor.LifeSustainingTreatmentOptionA.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{LifeSustainingTreatmentOption: actor.LifeSustainingTreatmentOptionA, Tasks: actor.DonorTasks{LifeSustainingTreatment: actor.TaskCompleted}}).
		Return(expectedError)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostLifeSustainingTreatmentWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *lifeSustainingTreatmentData) bool {
			return assert.Equal(t, validation.With("option", validation.SelectError{Label: "ifTheDonorGivesConsentToLifeSustainingTreatment"}), data.Errors)
		})).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadLifeSustainingTreatmentForm(t *testing.T) {
	form := url.Values{
		"option": {actor.LifeSustainingTreatmentOptionA.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readLifeSustainingTreatmentForm(r)

	assert.Equal(t, actor.LifeSustainingTreatmentOptionA, result.Option)
}

func TestLifeSustainingTreatmentFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lifeSustainingTreatmentForm
		errors validation.List
	}{
		"valid": {
			form: &lifeSustainingTreatmentForm{},
		},
		"invalid": {
			form: &lifeSustainingTreatmentForm{
				Error: expectedError,
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
