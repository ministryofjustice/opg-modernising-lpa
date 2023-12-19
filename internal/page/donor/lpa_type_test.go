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

func TestGetLpaType(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App:     testAppData,
			Form:    &lpaTypeForm{},
			Options: actor.LpaTypeValues,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &lpaTypeData{
			App: testAppData,
			Form: &lpaTypeForm{
				LpaType: actor.LpaTypePropertyAndAffairs,
			},
			Options: actor.LpaTypeValues,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Type: actor.LpaTypePropertyAndAffairs})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLpaType(t *testing.T) {
	testcases := map[actor.LpaType]*actor.DonorProvidedDetails{
		actor.LpaTypePropertyAndAffairs: {
			LpaID: "lpa-id",
			Type:  actor.LpaTypePropertyAndAffairs,
			Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted},
		},
		actor.LpaTypePersonalWelfare: {
			LpaID:               "lpa-id",
			Type:                actor.LpaTypePersonalWelfare,
			WhenCanTheLpaBeUsed: actor.CanBeUsedWhenCapacityLost,
			Tasks:               actor.DonorTasks{YourDetails: actor.TaskCompleted},
		},
	}

	for lpaType, donor := range testcases {
		t.Run(lpaType.String(), func(t *testing.T) {
			form := url.Values{
				"lpa-type": {lpaType.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.
				On("Put", r.Context(), donor).
				Return(nil)

			err := LpaType(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID:                          "lpa-id",
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostLpaTypeWhenNotChanged(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := LpaType(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
		Type:  actor.LpaTypePropertyAndAffairs,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostLpaTypeWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := LpaType(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostLpaTypeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.MatchedBy(func(data *lpaTypeData) bool {
			return assert.Equal(t, validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}), data.Errors)
		})).
		Return(nil)

	err := LpaType(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadLpaTypeForm(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readLpaTypeForm(r)

	assert.Equal(t, actor.LpaTypePropertyAndAffairs, result.LpaType)
}

func TestLpaTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lpaTypeForm
		errors validation.List
	}{
		"valid": {
			form: &lpaTypeForm{},
		},
		"invalid": {
			form: &lpaTypeForm{
				Error: expectedError,
			},
			errors: validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
