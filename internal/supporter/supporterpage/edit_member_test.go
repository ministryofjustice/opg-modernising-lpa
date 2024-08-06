package supporterpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEditMember(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=an-id", nil)

	member := &supporterdata.Member{
		ID:         "an-id",
		FirstNames: "a",
		LastName:   "b",
		Status:     supporterdata.StatusActive,
		Permission: supporterdata.PermissionAdmin,
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
				Permission:        supporterdata.PermissionAdmin,
				PermissionOptions: supporterdata.PermissionValues,
				Status:            supporterdata.StatusActive,
				StatusOptions:     supporterdata.StatusValues,
			},
			Member: member,
		}).
		Return(nil)

	err := EditMember(nil, template.Execute, memberStore)(testAppData, w, r, &supporterdata.Organisation{}, &supporterdata.Member{})
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

	err := EditMember(nil, nil, memberStore)(testAppData, w, r, &supporterdata.Organisation{}, &supporterdata.Member{})
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

	err := EditMember(nil, template.Execute, nil)(testAppData, w, r, &supporterdata.Organisation{}, &supporterdata.Member{ID: "an-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEditMember(t *testing.T) {
	testcases := map[string]struct {
		member           *supporterdata.Member
		expectedRedirect string
		expectedMember   *supporterdata.Member
		userPermission   supporterdata.Permission
		memberEmail      string
	}{
		"self": {
			member: &supporterdata.Member{
				ID:         "an-id",
				FirstNames: "a",
				LastName:   "b",
				Email:      "self@example.org",
				Status:     supporterdata.StatusActive,
				Permission: supporterdata.PermissionAdmin,
			},
			userPermission:   supporterdata.PermissionAdmin,
			memberEmail:      "self@example.org",
			expectedRedirect: supporter.PathManageTeamMembers.Format() + "?nameUpdated=c+d&selfUpdated=1",
			expectedMember: &supporterdata.Member{
				ID:         "an-id",
				FirstNames: "c",
				LastName:   "d",
				Email:      "self@example.org",
				Status:     supporterdata.StatusActive,
				Permission: supporterdata.PermissionAdmin,
			},
		},
		"non-admin": {
			member: &supporterdata.Member{
				ID:    "an-id",
				Email: "self@example.org",
			},
			userPermission:   supporterdata.PermissionNone,
			memberEmail:      "self@example.org",
			expectedRedirect: supporter.PathDashboard.Format() + "?nameUpdated=c+d&selfUpdated=1",
			expectedMember: &supporterdata.Member{
				ID:         "an-id",
				FirstNames: "c",
				LastName:   "d",
				Email:      "self@example.org",
				Status:     supporterdata.StatusActive,
				Permission: supporterdata.PermissionNone,
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

			err := EditMember(nil, nil, memberStore)(appcontext.Data{
				LoginSessionEmail: "self@example.org",
				SupporterData: &appcontext.SupporterData{
					Permission: tc.userPermission,
				},
			}, w, r, &supporterdata.Organisation{}, tc.member)
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
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id=an-id", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetByID(r.Context(), "an-id").
		Return(&supporterdata.Member{
			ID:         "member-id",
			FirstNames: "a",
			LastName:   "b",
			Email:      "team-member@example.org",
			Status:     supporterdata.StatusActive,
			Permission: supporterdata.PermissionNone,
		}, nil)
	memberStore.EXPECT().
		Put(r.Context(), &supporterdata.Member{
			ID:         "member-id",
			FirstNames: "c",
			LastName:   "d",
			Email:      "team-member@example.org",
			Status:     supporterdata.StatusSuspended,
			Permission: supporterdata.PermissionAdmin,
		}).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "member status changed", slog.String("member_id", "member-id"), slog.String("status_old", "active"), slog.String("status_new", "suspended"))
	logger.EXPECT().
		InfoContext(r.Context(), "member permission changed", slog.String("member_id", "member-id"), slog.String("permission_old", "none"), slog.String("permission_new", "admin"))

	err := EditMember(logger, nil, memberStore)(appcontext.Data{
		LoginSessionEmail: "self@example.org",
		SupporterData: &appcontext.SupporterData{
			Permission: supporterdata.PermissionAdmin,
		},
	}, w, r, &supporterdata.Organisation{}, &supporterdata.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathManageTeamMembers.Format()+"?nameUpdated=c+d&statusEmail=team-member%40example.org&statusUpdated=suspended", resp.Header.Get("Location"))
}

func TestPostEditMemberNoUpdate(t *testing.T) {
	testcases := map[string]struct {
		expectedRedirect string
		userPermission   supporterdata.Permission
		memberEmail      string
	}{
		"self": {
			userPermission:   supporterdata.PermissionAdmin,
			memberEmail:      "self@example.org",
			expectedRedirect: supporter.PathManageTeamMembers.Format() + "?",
		},
		"non-admin": {
			userPermission:   supporterdata.PermissionNone,
			memberEmail:      "self@example.org",
			expectedRedirect: supporter.PathDashboard.Format() + "?",
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

			err := EditMember(nil, nil, nil)(appcontext.Data{
				LoginSessionEmail: "self@example.org",
				SupporterData: &appcontext.SupporterData{
					Permission: tc.userPermission,
				},
			}, w, r, &supporterdata.Organisation{}, &supporterdata.Member{
				ID:         "an-id",
				FirstNames: "a",
				LastName:   "b",
				Email:      tc.memberEmail,
				Status:     supporterdata.StatusActive,
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
		Return(&supporterdata.Member{
			FirstNames: "a",
			LastName:   "b",
			Email:      "team-member@example.org",
			Status:     supporterdata.StatusActive,
			Permission: supporterdata.PermissionAdmin,
		}, nil)

	err := EditMember(nil, nil, memberStore)(appcontext.Data{
		LoginSessionEmail: "self@example.org",
		SupporterData: &appcontext.SupporterData{
			Permission: supporterdata.PermissionAdmin,
		},
	}, w, r, &supporterdata.Organisation{}, &supporterdata.Member{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathManageTeamMembers.Format()+"?", resp.Header.Get("Location"))
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
		Return(&supporterdata.Member{}, nil)

	memberStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EditMember(nil, nil, memberStore)(testAppData, w, r, &supporterdata.Organisation{}, &supporterdata.Member{})
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

	member := &supporterdata.Member{
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

	err := EditMember(nil, template.Execute, memberStore)(testAppData, w, r, nil, &supporterdata.Member{})
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
				assert.Equal(t, supporterdata.StatusSuspended, result.Status)
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
				Status:     supporterdata.StatusSuspended,
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
