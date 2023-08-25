package donor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSelectYourIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:  testAppData,
			Page: 2,
			Form: &selectYourIdentityOptionsForm{},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, nil, 2)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSelectYourIdentityOptionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:  testAppData,
			Form: &selectYourIdentityOptionsForm{Selected: identity.Passport},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, nil, 0)(testAppData, w, r, &page.Lpa{
		DonorIdentityOption: identity.Passport,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSelectYourIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := SelectYourIdentityOptions(template.Execute, nil, 0)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostSelectYourIdentityOptions(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			ID:                  "lpa-id",
			DonorIdentityOption: identity.Passport,
			Tasks: page.Tasks{
				ConfirmYourIdentityAndSign: actor.TaskInProgress,
			},
		}).
		Return(nil)

	err := SelectYourIdentityOptions(nil, donorStore, 0)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.YourChosenIdentityOptions.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostSelectYourIdentityOptionsNone(t *testing.T) {
	for pageIndex, nextPath := range map[int]page.LpaPath{
		0: page.Paths.SelectYourIdentityOptions1,
		1: page.Paths.SelectYourIdentityOptions2,
		2: page.Paths.TaskList,
	} {
		t.Run(fmt.Sprintf("Page%d", pageIndex), func(t *testing.T) {
			form := url.Values{
				"option": {"none"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := SelectYourIdentityOptions(nil, nil, pageIndex)(testAppData, w, r, &page.Lpa{ID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, nextPath.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostSelectYourIdentityOptionsWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := SelectYourIdentityOptions(nil, donorStore, 0)(testAppData, w, r, &page.Lpa{})

	assert.Equal(t, expectedError, err)
}

func TestPostSelectYourIdentityOptionsWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &selectYourIdentityOptionsData{
			App:    testAppData,
			Form:   &selectYourIdentityOptionsForm{},
			Errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		}).
		Return(nil)

	err := SelectYourIdentityOptions(template.Execute, nil, 0)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadSelectYourIdentityOptionsForm(t *testing.T) {
	form := url.Values{
		"option": {"passport"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readSelectYourIdentityOptionsForm(r)

	assert.Equal(t, identity.Passport, result.Selected)
	assert.False(t, result.None)
}

func TestSelectYourIdentityOptionsFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form      *selectYourIdentityOptionsForm
		errors    validation.List
		pageIndex int
	}{
		"valid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.EasyID,
			},
		},
		"none": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
				None:     true,
			},
		},
		"missing": {
			form:   &selectYourIdentityOptionsForm{},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
		"invalid": {
			form: &selectYourIdentityOptionsForm{
				Selected: identity.UnknownOption,
			},
			errors: validation.With("option", validation.SelectError{Label: "fromTheListedOptions"}),
		},
		"missing after first page": {
			form:      &selectYourIdentityOptionsForm{},
			errors:    validation.With("option", validation.SelectError{Label: "whichDocumentYouWillUse"}),
			pageIndex: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate(tc.pageIndex))
		})
	}
}
