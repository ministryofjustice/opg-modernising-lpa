package supporterpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterYourName(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterYourNameData{
			App:  testAppData,
			Form: &enterYourNameForm{},
		}).
		Return(nil)

	err := EnterYourName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterYourNameWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EnterYourName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterYourName(t *testing.T) {
	form := url.Values{"first-names": {"John"}, "last-name": {"Smith"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Create(r.Context(), "John", "Smith").
		Return(&supporterdata.Member{}, nil)

	err := EnterYourName(nil, memberStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterEnterOrganisationName.Format(), resp.Header.Get("Location"))
}

func TestPostEnterYourNameWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{"last-name": {"a"}}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataMatcher := func(t *testing.T, data *enterYourNameData) bool {
		return assert.Equal(t, validation.With("first-names", validation.EnterError{Label: "firstNames"}), data.Errors)
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *enterYourNameData) bool {
			return dataMatcher(t, data)
		})).
		Return(nil)

	err := EnterYourName(template.Execute, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterYourNameWhenMemberStoreErrors(t *testing.T) {
	form := url.Values{"first-names": {"a"}, "last-name": {"b"}}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Create(r.Context(), mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterYourName(nil, memberStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEnterYourNameForm(t *testing.T) {
	form := url.Values{
		"first-names": {"a"},
		"last-name":   {"b"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEnterYourNameForm(r)

	assert.Equal(t, "a", result.FirstNames)
	assert.Equal(t, "b", result.LastName)
}

func TestEnterYourNameFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterYourNameForm
		errors validation.List
	}{
		"valid": {
			form: &enterYourNameForm{
				FirstNames: "John",
				LastName:   "Smith",
			},
		},
		"missing": {
			form: &enterYourNameForm{},
			errors: validation.With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &enterYourNameForm{
				FirstNames: strings.Repeat("a", 54),
				LastName:   strings.Repeat("b", 62),
			},
			errors: validation.With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
