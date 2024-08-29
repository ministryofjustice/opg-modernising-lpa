package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
)

func TestGetRestrictions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &restrictionsData{
			App:   testAppData,
			Donor: &donordata.Provided{},
		}).
		Return(nil)

	err := Restrictions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRestrictionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &restrictionsData{
			App:   testAppData,
			Donor: &donordata.Provided{Restrictions: "blah"},
		}).
		Return(nil)

	err := Restrictions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Restrictions: "blah"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRestrictionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &restrictionsData{
			App:   testAppData,
			Donor: &donordata.Provided{},
		}).
		Return(expectedError)

	err := Restrictions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:        "lpa-id",
			Restrictions: "blah",
			Tasks:        donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted, Restrictions: task.StateCompleted},
		}).
		Return(nil)

	err := Restrictions(nil, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRestrictionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"restrictions": {"blah"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{Restrictions: "blah", Tasks: donordata.Tasks{Restrictions: task.StateCompleted}}).
		Return(expectedError)

	err := Restrictions(nil, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostRestrictionsWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"restrictions": {random.String(10001)},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &restrictionsData{
			App:    testAppData,
			Errors: validation.With("restrictions", validation.StringTooLongError{Label: "restrictions", Length: 10000}),
			Donor:  &donordata.Provided{},
		}).
		Return(nil)

	err := Restrictions(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
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
