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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whenCanTheLpaBeUsedData{
			App:     testAppData,
			Donor:   &actor.DonorProvidedDetails{},
			Form:    &whenCanTheLpaBeUsedForm{},
			Options: donordata.CanBeUsedWhenValues,
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whenCanTheLpaBeUsedData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{WhenCanTheLpaBeUsed: donordata.CanBeUsedWhenHasCapacity},
			Form: &whenCanTheLpaBeUsedForm{
				When: donordata.CanBeUsedWhenHasCapacity,
			},
			Options: donordata.CanBeUsedWhenValues,
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{WhenCanTheLpaBeUsed: donordata.CanBeUsedWhenHasCapacity})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWhenCanTheLpaBeUsedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	form := url.Values{
		"when": {donordata.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:               "lpa-id",
			WhenCanTheLpaBeUsed: donordata.CanBeUsedWhenHasCapacity,
			Tasks:               actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, WhenCanTheLpaBeUsed: actor.TaskCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"when": {donordata.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{WhenCanTheLpaBeUsed: donordata.CanBeUsedWhenHasCapacity, Tasks: actor.DonorTasks{WhenCanTheLpaBeUsed: actor.TaskCompleted}}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whenCanTheLpaBeUsedData) bool {
			return assert.Equal(t, validation.With("when", validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}), data.Errors)
		})).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadWhenCanTheLpaBeUsedForm(t *testing.T) {
	form := url.Values{
		"when": {donordata.CanBeUsedWhenHasCapacity.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhenCanTheLpaBeUsedForm(r)

	assert.Equal(t, donordata.CanBeUsedWhenHasCapacity, result.When)
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
