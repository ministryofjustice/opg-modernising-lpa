package supporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInviteMember(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &inviteMemberData{
			App:     testAppData,
			Form:    &inviteMemberForm{},
			Options: actor.PermissionValues,
		}).
		Return(nil)

	err := InviteMember(template.Execute, nil, nil, nil, "http://base")(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetInviteMemberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := InviteMember(template.Execute, nil, nil, nil, "http://base")(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostInviteMember(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "inviter@example.com"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisation := &actor.Organisation{Name: "My organisation"}

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateMemberInvite(r.Context(), organisation, "a", "b", "email@example.com", "abcde", actor.Admin).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), "email@example.com", notify.OrganisationMemberInviteEmail{
			OrganisationName:      "My organisation",
			InviterEmail:          "inviter@example.com",
			InviteCode:            "abcde",
			JoinAnOrganisationURL: "http://base" + page.Paths.Supporter.Start.Format(),
		}).
		Return(nil)

	err := InviteMember(nil, organisationStore, notifyClient, func(int) string { return "abcde" }, "http://base")(testAppData, w, r, organisation)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.InviteMemberConfirmation.Format()+"?email=email%40example.com", resp.Header.Get("Location"))
}

func TestPostInviteMemberWhenValidationError(t *testing.T) {
	form := url.Values{
		"email":       {"what"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &inviteMemberData{
			App:    testAppData,
			Errors: validation.With("email", validation.EmailError{Label: "email"}),
			Form: &inviteMemberForm{
				FirstNames: "a",
				LastName:   "b",
				Email:      "what",
				Permission: actor.Admin,
			},
			Options: actor.PermissionValues,
		}).
		Return(nil)

	err := InviteMember(template.Execute, nil, nil, nil, "http://base")(testAppData, w, r, nil)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostInviteMemberWhenCreateMemberInviteErrors(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"none"},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "inviter@example.com"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateMemberInvite(r.Context(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := InviteMember(nil, organisationStore, nil, func(int) string { return "abcde" }, "http://base")(testAppData, w, r, &actor.Organisation{})
	assert.Equal(t, expectedError, err)
}

func TestPostInviteMemberWhenNotifySendErrors(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "inviter@example.com"})
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		CreateMemberInvite(r.Context(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendEmail(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	err := InviteMember(nil, organisationStore, notifyClient, func(int) string { return "abcde" }, "http://base")(testAppData, w, r, &actor.Organisation{})
	assert.Equal(t, expectedError, err)
}

func TestReadInviteMemberForm(t *testing.T) {
	form := url.Values{
		"email":       {"email@example.com"},
		"first-names": {"a"},
		"last-name":   {"b"},
		"permission":  {"admin"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readInviteMemberForm(r)

	assert.Equal(t, "email@example.com", result.Email)
	assert.Equal(t, "a", result.FirstNames)
	assert.Equal(t, "b", result.LastName)
	assert.Equal(t, actor.Admin, result.Permission)
}

func TestInviteMemberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *inviteMemberForm
		errors validation.List
	}{
		"valid": {
			form: &inviteMemberForm{
				Email:      "email@example.com",
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.None,
			},
		},
		"missing": {
			form: &inviteMemberForm{},
			errors: validation.
				With("first-names", validation.EnterError{Label: "firstNames"}).
				With("last-name", validation.EnterError{Label: "lastName"}).
				With("email", validation.EnterError{Label: "email"}),
		},
		"invalid": {
			form: &inviteMemberForm{
				Email:      "what",
				FirstNames: "a",
				LastName:   "b",
				Permission: actor.None,
			},
			errors: validation.With("email", validation.EmailError{Label: "email"}),
		},
		"too long": {
			form: &inviteMemberForm{
				Email:      "email@example.com",
				FirstNames: strings.Repeat("x", 54),
				LastName:   strings.Repeat("x", 62),
				Permission: actor.None,
			},
			errors: validation.
				With("first-names", validation.StringTooLongError{Label: "firstNames", Length: 53}).
				With("last-name", validation.StringTooLongError{Label: "lastName", Length: 61}),
		},
		"permission error": {
			form: &inviteMemberForm{
				Email:      "email@example.com",
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
