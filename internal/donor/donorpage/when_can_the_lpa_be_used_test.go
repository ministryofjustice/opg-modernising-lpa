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

func TestGetWhenCanTheLpaBeUsed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &whenCanTheLpaBeUsedData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  form.NewEmptySelectForm[lpadata.CanBeUsedWhen](lpadata.CanBeUsedWhenValues, "whenYourAttorneysCanUseYourLpa"),
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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
			Donor: &donordata.Provided{WhenCanTheLpaBeUsed: lpadata.CanBeUsedWhenHasCapacity},
			Form:  form.NewSelectForm(lpadata.CanBeUsedWhenHasCapacity, lpadata.CanBeUsedWhenValues, "whenYourAttorneysCanUseYourLpa"),
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &donordata.Provided{WhenCanTheLpaBeUsed: lpadata.CanBeUsedWhenHasCapacity})
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

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostWhenCanTheLpaBeUsed(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {lpadata.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:               "lpa-id",
			WhenCanTheLpaBeUsed: lpadata.CanBeUsedWhenHasCapacity,
			Tasks:               donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted, WhenCanTheLpaBeUsed: task.StateCompleted},
		}).
		Return(nil)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostWhenCanTheLpaBeUsedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		form.FieldNames.Select: {lpadata.CanBeUsedWhenHasCapacity.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{WhenCanTheLpaBeUsed: lpadata.CanBeUsedWhenHasCapacity, Tasks: donordata.Tasks{WhenCanTheLpaBeUsed: task.StateCompleted}}).
		Return(expectedError)

	err := WhenCanTheLpaBeUsed(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostWhenCanTheLpaBeUsedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *whenCanTheLpaBeUsedData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.Select, validation.SelectError{Label: "whenYourAttorneysCanUseYourLpa"}), data.Errors)
		})).
		Return(nil)

	err := WhenCanTheLpaBeUsed(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
