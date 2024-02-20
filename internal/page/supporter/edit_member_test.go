package supporter

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

func TestGetEditMember(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	member := &actor.Member{
		ID:         "an-id",
		FirstNames: "a",
		LastName:   "b",
		Permission: actor.Admin,
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Member(r.Context(), "an-id").
		Return(member, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &editMemberData{
			App: testAppData,
			Form: &editMemberForm{
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.Admin,
			},
			Options: actor.PermissionValues,
			Member:  member,
		}).
		Return(nil)

	err := EditMember(template.Execute, memberStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEditMemberWhenOrganisationStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Member(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := EditMember(nil, memberStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEditMemberWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Member(mock.Anything, mock.Anything).
		Return(&actor.Member{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EditMember(template.Execute, memberStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditMember(t *testing.T) {
	testcases := map[string]struct {
		form           url.Values
		expectedQuery  string
		expectedMember *actor.Member
	}{
		"name updated": {
			form: url.Values{
				"first-names": {"c"},
				"last-name":   {"d"},
				"permission":  {"admin"},
			},
			expectedQuery: "?nameUpdated=c+d",
			expectedMember: &actor.Member{
				FirstNames: "c",
				LastName:   "d",
				Permission: actor.Admin,
			},
		},
		"no updates": {
			form: url.Values{
				"first-names": {"a"},
				"last-name":   {"b"},
				"permission":  {"admin"},
			},
			expectedQuery: "?",
			expectedMember: &actor.Member{
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.Admin,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				Member(r.Context(), "an-id").
				Return(&actor.Member{
					FirstNames: "a",
					LastName:   "b",
					Permission: actor.Admin,
				}, nil)

			memberStore.EXPECT().
				Put(r.Context(), tc.expectedMember).
				Return(nil)

			err := EditMember(nil, memberStore)(testAppData, w, r, &actor.Organisation{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Supporter.ManageTeamMembers.Format()+tc.expectedQuery, resp.Header.Get("Location"))
		})
	}
}

func TestPostEditMemberWhenOrganisationStorePutError(t *testing.T) {
	form := url.Values{
		"first-names": {"c"},
		"last-name":   {"d"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Member(mock.Anything, mock.Anything).
		Return(&actor.Member{}, nil)

	memberStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EditMember(nil, memberStore)(testAppData, w, r, &actor.Organisation{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditMemberWhenValidationError(t *testing.T) {
	form := url.Values{
		"first-names": {""},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	member := &actor.Member{
		ID:         "an-id",
		FirstNames: "a",
		LastName:   "b",
		Permission: actor.Admin,
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		Member(r.Context(), "an-id").
		Return(member, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &editMemberData{
			App:    testAppData,
			Errors: validation.With("first-names", validation.EnterError{Label: "firstNames"}),
			Form: &editMemberForm{
				FirstNames: "",
				LastName:   "b",
				Permission: actor.Admin,
			},
			Options: actor.PermissionValues,
			Member:  member,
		}).
		Return(nil)

	err := EditMember(template.Execute, memberStore)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEditMemberForm(t *testing.T) {
	form := url.Values{
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readEditMemberForm(r)

	assert.Equal(t, "a", result.FirstNames)
	assert.Equal(t, "b", result.LastName)
	assert.Equal(t, actor.Admin, result.Permission)
}

func TestEditMemberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *editMemberForm
		errors validation.List
	}{
		"valid": {
			form: &editMemberForm{
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.None,
			},
		},
		"missing": {
			form: &editMemberForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}),
		},
		"too long": {
			form: &editMemberForm{
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Permission: actor.None,
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"permission error": {
			form: &editMemberForm{
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.Permission(99),
			},
			errors: validation.
				With("permission", validation.SelectError{Label: "makeThisPersonAnAdmin"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
