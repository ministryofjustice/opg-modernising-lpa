package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testNow   = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn = func() time.Time { return testNow }
)

func TestLoginCallbackNoOrganisation(t *testing.T) {
	loginSession := &sesh.LoginSession{
		IDToken: "id-token",
		Sub:     "supporter-random",
		Email:   "name@example.com",
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: "/redirect",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, loginSession).
		Return(nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembersByEmail(mock.Anything).
		Return([]*actor.MemberInvite{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
		Return(&actor.Organisation{}, dynamo.NotFoundError{})

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.EnterOrganisationName.Format(), resp.Header.Get("Location"))
}

func TestLoginCallbackErrorGettingOrganisation(t *testing.T) {
	loginSession := &sesh.LoginSession{
		IDToken: "id-token",
		Sub:     "supporter-random",
		Email:   "name@example.com",
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: "/redirect",
		}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, expectedError)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenNoOrganisationAndSetLoginSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), mock.Anything, mock.Anything).
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: "/redirect",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, dynamo.NotFoundError{})

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)

	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackIsOrganisationMember(t *testing.T) {
	testcases := map[string]struct {
		loginSessionEmail   string
		existingMemberEmail string
	}{
		"with same email": {
			loginSessionEmail:   "a@example.org",
			existingMemberEmail: "a@example.org",
		},
		"with new email": {
			loginSessionEmail:   "b@example.org",
			existingMemberEmail: "a@example.org",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

			client := newMockOneLoginClient(t)
			client.EXPECT().
				Exchange(r.Context(), "auth-code", "my-nonce").
				Return("id-token", "a JWT", nil)
			client.EXPECT().
				UserInfo(r.Context(), "a JWT").
				Return(onelogin.UserInfo{Sub: "random", Email: tc.loginSessionEmail}, nil)

			loginSession := &sesh.LoginSession{
				IDToken:          "id-token",
				Sub:              "supporter-random",
				Email:            tc.loginSessionEmail,
				OrganisationID:   "org-id",
				OrganisationName: "org name",
			}

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				OneLogin(r).
				Return(&sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				}, nil)
			sessionStore.EXPECT().
				SetLogin(r, w, loginSession).
				Return(nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email, OrganisationID: "org-id"})).
				Return(&actor.Member{Email: tc.existingMemberEmail}, nil)

			memberStore.EXPECT().
				Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email, OrganisationID: "org-id"}), &actor.Member{Email: tc.loginSessionEmail, LastLoggedInAt: testNow}).
				Return(nil)

			organisationStore := newMockOrganisationStore(t)
			organisationStore.EXPECT().
				Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
				Return(&actor.Organisation{ID: "org-id", Name: "org name"}, nil)

			err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)

			assert.Nil(t, err)
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Supporter.Dashboard.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestLoginCallbackIsOrganisationMemberErrors(t *testing.T) {
	testcases := map[string]struct {
		memberError    error
		putMemberError error
	}{
		"Get error": {
			memberError: expectedError,
		},
		"Put error": {
			putMemberError: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

			client := newMockOneLoginClient(t)
			client.EXPECT().
				Exchange(r.Context(), mock.Anything, mock.Anything).
				Return("id-token", "a JWT", nil)
			client.EXPECT().
				UserInfo(mock.Anything, mock.Anything).
				Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

			loginSession := &sesh.LoginSession{
				IDToken:          "id-token",
				Sub:              "supporter-random",
				Email:            "name@example.com",
				OrganisationID:   "org-id",
				OrganisationName: "org name",
			}

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				OneLogin(r).
				Return(&sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				}, nil)
			sessionStore.EXPECT().
				SetLogin(mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				Get(mock.Anything).
				Return(&actor.Member{Email: loginSession.Email}, tc.memberError)

			if tc.putMemberError != nil {
				memberStore.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(tc.putMemberError)
			}

			organisationStore := newMockOrganisationStore(t)
			organisationStore.EXPECT().
				Get(mock.Anything).
				Return(&actor.Organisation{ID: "org-id", Name: "org name"}, nil)

			err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)

			assert.Equal(t, expectedError, err)
			resp := w.Result()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestLoginCallbackWhenEmailHasInvite(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: "/redirect",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{
			IDToken: "id-token",
			Sub:     "supporter-random",
			Email:   "name@example.com",
		}).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, dynamo.NotFoundError{})

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembersByEmail(mock.Anything).
		Return([]*actor.MemberInvite{{}}, nil)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)

	assert.Nil(t, err)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.EnterReferenceNumber.Format(), resp.Header.Get("Location"))
}

func TestLoginCallbackWhenEmailHasInviteWhenInvitedMembersByEmailError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: "/redirect",
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{
			IDToken: "id-token",
			Sub:     "supporter-random",
			Email:   "name@example.com",
		}).
		Return(nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, dynamo.NotFoundError{})

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMembersByEmail(mock.Anything).
		Return([]*actor.MemberInvite{{}}, expectedError)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)

	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(nil, expectedError)

	err := LoginCallback(nil, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenExchangeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("", "", expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: page.Paths.Supporter.LoginCallback.Format()}, nil)

	err := LoginCallback(client, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: page.Paths.Supporter.LoginCallback.Format()}, nil)

	err := LoginCallback(client, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: page.Paths.Supporter.LoginCallback.Format(),
		}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackWhenOrganisationIsDeleted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "a@example.org"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(&sesh.OneLoginSession{
			State:    "my-state",
			Nonce:    "my-nonce",
			Locale:   "en",
			Redirect: page.Paths.Supporter.LoginCallback.Format(),
		}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{DeletedAt: time.Now()}, nil)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, nil)(page.AppData{}, w, r)

	assert.Nil(t, err)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Start.Format(), resp.Header.Get("Location"))
}
