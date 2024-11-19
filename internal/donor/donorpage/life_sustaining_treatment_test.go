package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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
			App:  testAppData,
			Form: form.NewEmptySelectForm[lpadata.LifeSustainingTreatment](lpadata.LifeSustainingTreatmentValues, "ifTheDonorGivesConsentToLifeSustainingTreatment"),
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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
			App:  testAppData,
			Form: form.NewSelectForm(lpadata.LifeSustainingTreatmentOptionA, lpadata.LifeSustainingTreatmentValues, "ifTheDonorGivesConsentToLifeSustainingTreatment"),
		}).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &donordata.Provided{LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA})
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

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLifeSustainingTreatment(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {lpadata.LifeSustainingTreatmentOptionA.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:                         "lpa-id",
			LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA,
			Tasks:                         donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted, LifeSustainingTreatment: task.StateCompleted},
		}).
		Return(nil)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostLifeSustainingTreatmentWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {lpadata.LifeSustainingTreatmentOptionA.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{LifeSustainingTreatmentOption: lpadata.LifeSustainingTreatmentOptionA, Tasks: donordata.Tasks{LifeSustainingTreatment: task.StateCompleted}}).
		Return(expectedError)

	err := LifeSustainingTreatment(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostLifeSustainingTreatmentWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *lifeSustainingTreatmentData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "ifTheDonorGivesConsentToLifeSustainingTreatment"}), data.Errors)
		})).
		Return(nil)

	err := LifeSustainingTreatment(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
