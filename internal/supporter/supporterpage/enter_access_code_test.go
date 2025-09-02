package supporterpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/invitecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterAccessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterAccessCode{
			App: testAppData,
			Form: &enterAccessCodeForm{
				FieldName: form.FieldNames.AccessCode,
			},
		}).
		Return(nil)

	err := EnterAccessCode(nil, template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterAccessCodeWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterAccessCode{
			App: testAppData,
			Form: &enterAccessCodeForm{
				FieldName: form.FieldNames.AccessCode,
			},
		}).
		Return(expectedError)

	err := EnterAccessCode(nil, template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCode(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"abcd1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invite := &supporterdata.MemberInvite{
		InviteCode:       invitecode.HashedFromString("abcd1234"),
		OrganisationID:   "org-id",
		OrganisationName: "org name",
		CreatedAt:        time.Now().Add(-47 * time.Hour),
		Email:            "a@example.org",
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(r.Context()).
		Return(invite, nil)

	memberStore.EXPECT().
		CreateFromInvite(r.Context(), invite).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub", OrganisationID: "org-id", OrganisationName: "org name"}).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "member invite redeemed", slog.String("organisation_id", "org-id"))

	err := EnterAccessCode(logger, nil, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, supporter.PathDashboard.Format(), resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeWhenIncorrectAccessCode(t *testing.T) {
	f := url.Values{form.FieldNames.AccessCode: {"not-match"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(r.Context()).
		Return(&supporterdata.MemberInvite{InviteCode: invitecode.HashedFromString("notmatch123"), OrganisationID: "org-id"}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterAccessCode{
			App: testAppData,
			Form: &enterAccessCodeForm{
				FieldName:     form.FieldNames.AccessCode,
				AccessCode:    "notmatch",
				AccessCodeRaw: "not-match",
			},
			Errors: validation.With(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"}),
		}).
		Return(nil)

	err := EnterAccessCode(nil, template.Execute, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeWhenInviteExpired(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"match-123"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(r.Context()).
		Return(&supporterdata.MemberInvite{
			InviteCode:     invitecode.HashedFromString("match123"),
			OrganisationID: "org-id",
			CreatedAt:      time.Now().Add(-49 * time.Hour),
		}, nil)

	err := EnterAccessCode(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterInviteExpired.Format(), resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeWhenMemberStoreInvitedMemberError(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"abcd1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&supporterdata.MemberInvite{}, expectedError)

	err := EnterAccessCode(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeWhenMemberStoreCreateError(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"abcd1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&supporterdata.MemberInvite{
			InviteCode:     invitecode.HashedFromString("abcd1234"),
			OrganisationID: "org-id",
			CreatedAt:      time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeWhenSessionGetError(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"abcd1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&supporterdata.MemberInvite{
			InviteCode:     invitecode.HashedFromString("abcd1234"),
			OrganisationID: "org-id",
			CreatedAt:      time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	err := EnterAccessCode(nil, nil, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathSupporterStart.Format(), resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeWhenSessionSaveError(t *testing.T) {
	form := url.Values{form.FieldNames.AccessCode: {"abcd1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&supporterdata.MemberInvite{
			InviteCode:     invitecode.HashedFromString("abcd1234"),
			OrganisationID: "org-id",
			CreatedAt:      time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := EnterAccessCode(logger, nil, memberStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeWhenValidationError(t *testing.T) {
	f := url.Values{form.FieldNames.AccessCode: {""}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterAccessCode{
			App: testAppData,
			Form: &enterAccessCodeForm{
				FieldName: form.FieldNames.AccessCode,
			},
			Errors: validation.With(form.FieldNames.AccessCode, validation.EnterError{Label: "yourAccessCode"}),
		}).
		Return(nil)

	err := EnterAccessCode(nil, template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAccessCodeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *enterAccessCodeForm
		errors validation.List
	}{
		"valid": {
			form:   &enterAccessCodeForm{AccessCode: "abcd1234"},
			errors: nil,
		},
		"too short": {
			form: &enterAccessCodeForm{AccessCode: "abcd123"},
			errors: validation.With(form.FieldNames.AccessCode, validation.StringLengthError{
				Label:  "theAccessCodeYouEnter",
				Length: 8,
			}),
		},
		"too long": {
			form: &enterAccessCodeForm{AccessCode: "abcd12345"},
			errors: validation.With(form.FieldNames.AccessCode, validation.StringLengthError{
				Label:  "theAccessCodeYouEnter",
				Length: 8,
			}),
		},
		"empty": {
			form: &enterAccessCodeForm{},
			errors: validation.With(form.FieldNames.AccessCode, validation.EnterError{
				Label: "yourAccessCode",
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
