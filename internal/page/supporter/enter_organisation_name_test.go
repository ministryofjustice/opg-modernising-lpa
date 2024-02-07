package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterOrganisationName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterOrganisationNameData{
			App:  testAppData,
			Form: &organisationNameForm{},
		}).
		Return(nil)

	err := EnterOrganisationName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterOrganisationNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterOrganisationName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationName(t *testing.T) {
	form := url.Values{"name": {"My organisation"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), "My organisation").
		Return(nil)

	err := EnterOrganisationName(nil, organisationStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.OrganisationCreated.Format(), resp.Header.Get("Location"))
}

func TestPostEnterOrganisationNameWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *enterOrganisationNameData) bool {
		return assert.Equal(t, validation.With("name", validation.EnterError{Label: "fullOrganisationOrCompanyName"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterOrganisationNameData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterOrganisationName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterOrganisationNameWhenOrganisationStoreErrors(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Create(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterOrganisationName(nil, organisationStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEnterOrganisationNameForm(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readOrganisationNameForm(r, "x")

	assert.Equal(t, "My name", result.Name)
}

func TestEnterOrganisationNameFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *organisationNameForm
		errors validation.List
	}{
		"valid": {
			form: &organisationNameForm{
				Name: "My name",
			},
		},
		"missing": {
			form:   &organisationNameForm{Label: "xyz"},
			errors: validation.With("name", validation.EnterError{Label: "xyz"}),
		},
		"too long": {
			form: &organisationNameForm{
				Name:  strings.Repeat("a", 101),
				Label: "xyz",
			},
			errors: validation.With("name", validation.StringTooLongError{Label: "xyz", Length: 100}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
