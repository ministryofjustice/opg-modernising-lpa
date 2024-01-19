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

func TestGetEnterGroupName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterGroupNameData{
			App:  testAppData,
			Form: &enterGroupNameForm{},
		}).
		Return(nil)

	err := EnterGroupName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterGroupNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterGroupName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterGroupName(t *testing.T) {
	form := url.Values{"name": {"My group"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	groupStore := newMockGroupStore(t)
	groupStore.EXPECT().
		Create(r.Context(), "My group").
		Return(nil)

	err := EnterGroupName(nil, groupStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.GroupCreated.Format(), resp.Header.Get("Location"))
}

func TestPostEnterGroupNameWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *enterGroupNameData) bool {
		return assert.Equal(t, validation.With("name", validation.EnterError{Label: "fullOrganisationOrCompanyName"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterGroupNameData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterGroupName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterGroupNameWhenGroupStoreErrors(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	groupStore := newMockGroupStore(t)
	groupStore.EXPECT().
		Create(r.Context(), mock.Anything).
		Return(expectedError)

	err := EnterGroupName(nil, groupStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEnterGroupNameForm(t *testing.T) {
	form := url.Values{
		"name": {"My name"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterGroupNameForm(r)

	assert.Equal(t, "My name", result.Name)
}

func TestEnterGroupNameFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterGroupNameForm
		errors validation.List
	}{
		"valid": {
			form: &enterGroupNameForm{
				Name: "My name",
			},
		},
		"missing": {
			form:   &enterGroupNameForm{},
			errors: validation.With("name", validation.EnterError{Label: "fullOrganisationOrCompanyName"}),
		},
		"too long": {
			form: &enterGroupNameForm{
				Name: strings.Repeat("a", 101),
			},
			errors: validation.With("name", validation.StringTooLongError{Label: "fullOrganisationOrCompanyName", Length: 100}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
