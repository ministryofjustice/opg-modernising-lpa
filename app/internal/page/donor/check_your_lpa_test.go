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

func TestGetCheckYourLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourLpaData{
			App:  testAppData,
			Form: &checkYourLpaForm{},
			Lpa:  &page.Lpa{},
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetCheckYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{
		Checked:      true,
		HappyToShare: true,
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &checkYourLpaData{
			App: testAppData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				Checked: true,
				Happy:   true,
			},
		}).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostCheckYourLpa(t *testing.T) {
	form := url.Values{
		"checked": {"1"},
		"happy":   {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpa := &page.Lpa{
		Checked:      false,
		HappyToShare: false,
		Tasks:        page.Tasks{CheckYourLpa: actor.TaskInProgress},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Checked:      true,
			HappyToShare: true,
			Tasks:        page.Tasks{CheckYourLpa: actor.TaskCompleted},
		}).
		Return(nil)

	err := CheckYourLpa(nil, donorStore)(testAppData, w, r, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.TaskList, resp.Header.Get("Location"))
}

func TestPostCheckYourLpaWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"checked": {"1"},
		"happy":   {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Checked:      true,
			HappyToShare: true,
			Tasks:        page.Tasks{CheckYourLpa: actor.TaskCompleted},
		}).
		Return(expectedError)

	err := CheckYourLpa(nil, donorStore)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostCheckYourLpaWhenValidationErrors(t *testing.T) {
	form := url.Values{
		"checked": {"1"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *checkYourLpaData) bool {
			return assert.Equal(t, validation.With("happy", validation.SelectError{Label: "happyToShareLpa"}), data.Errors)
		})).
		Return(nil)

	err := CheckYourLpa(template.Execute, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadCheckYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"checked": {" 1   "},
		"happy":   {" 0"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readCheckYourLpaForm(r)

	assert.Equal(true, result.Checked)
	assert.Equal(false, result.Happy)
}

func TestCheckYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *checkYourLpaForm
		errors validation.List
	}{
		"valid": {
			form: &checkYourLpaForm{
				Happy:   true,
				Checked: true,
			},
		},
		"invalid-all": {
			form: &checkYourLpaForm{
				Happy:   false,
				Checked: false,
			},
			errors: validation.
				With("checked", validation.SelectError{Label: "checkedLpa"}).
				With("happy", validation.SelectError{Label: "happyToShareLpa"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
