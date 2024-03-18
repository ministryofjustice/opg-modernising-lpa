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
		Status:     actor.StatusActive,
		Permission: actor.PermissionAdmin,
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), "an-id").
		Return(member, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &editMemberData{
			App: testAppData,
			Form: &editMemberForm{
				FirstNames:        "a",
				LastName:          "b",
				Permission:        actor.PermissionAdmin,
				PermissionOptions: actor.PermissionValues,
				Status:            actor.StatusActive,
				StatusOptions:     actor.StatusValues,
			},
			Member: member,
		}).
		Return(nil)

	err := EditMember(template.Execute, memberStore)(testAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEditMemberWhenMemberStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), mock.Anything).
		Return(nil, expectedError)

	err := EditMember(nil, memberStore)(testAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEditMemberWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := EditMember(template.Execute, nil)(testAppData, w, r, &actor.Organisation{}, &actor.Member{ID: "an-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditMember(t *testing.T) {
	testcases := map[string]struct {
		member           *actor.Member
		expectedRedirect string
		expectedMember   *actor.Member
		userPermission   actor.Permission
		memberEmail      string
	}{
		"self": {
			member: &actor.Member{
				ID:         "an-id",
				FirstNames: "a",
				LastName:   "b",
				Email:      "self@example.org",
				Status:     actor.StatusActive,
				Permission: actor.PermissionAdmin,
			},
			userPermission:   actor.PermissionAdmin,
			memberEmail:      "self@example.org",
			expectedRedirect: page.Paths.Supporter.ManageTeamMembers.Format() + "?nameUpdated=c+d&selfUpdated=1",
			expectedMember: &actor.Member{
				ID:         "an-id",
				FirstNames: "c",
				LastName:   "d",
				Email:      "self@example.org",
				Status:     actor.StatusActive,
				Permission: actor.PermissionAdmin,
			},
		},
		"non-admin": {
			member: &actor.Member{
				ID:    "an-id",
				Email: "self@example.org",
			},
			userPermission:   actor.PermissionNone,
			memberEmail:      "self@example.org",
			expectedRedirect: page.Paths.Supporter.Dashboard.Format() + "?nameUpdated=c+d&selfUpdated=1",
			expectedMember: &actor.Member{
				ID:         "an-id",
				FirstNames: "c",
				LastName:   "d",
				Email:      "self@example.org",
				Status:     actor.StatusActive,
				Permission: actor.PermissionNone,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"first-names": {"c"},
				"last-name":   {"d"},
				"status":      {"suspended"},
				"permission":  {},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				Put(r.Context(), tc.expectedMember).
				Return(nil)

			err := EditMember(nil, memberStore)(page.AppData{
				LoginSessionEmail: "self@example.org",
				Permission:        tc.userPermission,
			}, w, r, &actor.Organisation{}, tc.member)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostEditMemberWhenOtherMember(t *testing.T) {
	form := url.Values{
		"first-names": {"c"},
		"last-name":   {"d"},
		"status":      {"suspended"},
		"permission":  {},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), "an-id").
		Return(&actor.Member{
			FirstNames: "a",
			LastName:   "b",
			Email:      "team-member@example.org",
			Status:     actor.StatusActive,
			Permission: actor.PermissionNone,
		}, nil)
	memberStore.EXPECT().
		Put(r.Context(), &actor.Member{
			FirstNames: "c",
			LastName:   "d",
			Email:      "team-member@example.org",
			Status:     actor.StatusSuspended,
			Permission: actor.PermissionNone,
		}).
		Return(nil)

	err := EditMember(nil, memberStore)(page.AppData{
		LoginSessionEmail: "self@example.org",
		Permission:        actor.PermissionAdmin,
	}, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ManageTeamMembers.Format()+"?nameUpdated=c+d&statusEmail=team-member%40example.org&statusUpdated=suspended", resp.Header.Get("Location"))
}

func TestPostEditMemberNoUpdate(t *testing.T) {
	testcases := map[string]struct {
		expectedRedirect string
		userPermission   actor.Permission
		memberEmail      string
	}{
		"self": {
			userPermission:   actor.PermissionAdmin,
			memberEmail:      "self@example.org",
			expectedRedirect: page.Paths.Supporter.ManageTeamMembers.Format() + "?",
		},
		"non-admin": {
			userPermission:   actor.PermissionNone,
			memberEmail:      "self@example.org",
			expectedRedirect: page.Paths.Supporter.Dashboard.Format() + "?",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"first-names": {"a"},
				"last-name":   {"b"},
				"status":      {"active"},
				"permission":  {"admin"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := EditMember(nil, nil)(page.AppData{
				LoginSessionEmail: "self@example.org",
				Permission:        tc.userPermission,
			}, w, r, &actor.Organisation{}, &actor.Member{
				ID:         "an-id",
				FirstNames: "a",
				LastName:   "b",
				Email:      tc.memberEmail,
				Status:     actor.StatusActive,
				Permission: tc.userPermission,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostEditMemberNoUpdateWhenOtherMember(t *testing.T) {
	form := url.Values{
		"first-names": {"a"},
		"last-name":   {"b"},
		"status":      {"active"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), "an-id").
		Return(&actor.Member{
			FirstNames: "a",
			LastName:   "b",
			Email:      "team-member@example.org",
			Status:     actor.StatusActive,
			Permission: actor.PermissionAdmin,
		}, nil)

	err := EditMember(nil, memberStore)(page.AppData{
		LoginSessionEmail: "self@example.org",
		Permission:        actor.PermissionAdmin,
	}, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.ManageTeamMembers.Format()+"?", resp.Header.Get("Location"))
}

func TestPostEditMemberWhenOrganisationStorePutError(t *testing.T) {
	form := url.Values{
		"first-names": {"c"},
		"last-name":   {"d"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		Return(&actor.Member{}, nil)

	memberStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EditMember(nil, memberStore)(testAppData, w, r, &actor.Organisation{}, &actor.Member{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditMemberWhenValidationError(t *testing.T) {
	form := url.Values{
		"first-names": {""},
		"last-name":   {"b"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	member := &actor.Member{
		ID:         "an-id",
		FirstNames: "a",
		LastName:   "b",
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), "an-id").
		Return(member, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &editMemberData{
			App:    testAppData,
			Errors: validation.With("first-names", validation.EnterError{Label: "firstNames"}),
			Form: &editMemberForm{
				FirstNames: "",
				LastName:   "b",
			},
			Member: member,
		}).
		Return(nil)

	err := EditMember(template.Execute, memberStore)(testAppData, w, r, nil, &actor.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadEditMemberForm(t *testing.T) {
	testcases := map[string]struct {
		isAdmin       bool
		isEditingSelf bool
		canEditAll    bool
		form          url.Values
	}{
		"can edit all": {
			canEditAll: true,
			form: url.Values{
				"first-names": {"a"},
				"last-name":   {"b"},
				"status":      {"suspended"},
			},
		},
		"can only edit name": {
			form: url.Values{
				"first-names": {"a"},
				"last-name":   {"b"},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			result := readEditMemberForm(r, tc.canEditAll)

			assert.Equal(t, "a", result.FirstNames)
			assert.Equal(t, "b", result.LastName)

			if tc.isAdmin && !tc.isEditingSelf {
				assert.Equal(t, actor.StatusSuspended, result.Status)
			}
		})
	}
}

func TestEditMemberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *editMemberForm
		errors validation.List
	}{
		"valid - non-admin": {
			form: &editMemberForm{
				FirstNames: "a",
				LastName:   "b",
			},
		},
		"valid - admin": {
			form: &editMemberForm{
				FirstNames: "a",
				LastName:   "b",
				Status:     actor.StatusSuspended,
				canEditAll: true,
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
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"unsupported status option": {
			form: &editMemberForm{
				FirstNames:  "a",
				LastName:    "b",
				StatusError: expectedError,
				canEditAll:  true,
			},
			errors: validation.With("status", validation.SelectError{Label: "status"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
