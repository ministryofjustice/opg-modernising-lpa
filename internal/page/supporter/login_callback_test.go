package supporter

import (
	"errors"
	"log/slog"
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

func TestLoginCallback(t *testing.T) {
	testcases := map[string]struct {
		invites  []*actor.MemberInvite
		redirect page.Path
	}{
		"no invite": {
			redirect: page.Paths.Supporter.EnterYourName,
		},
		"has invite": {
			invites:  []*actor.MemberInvite{{}},
			redirect: page.Paths.Supporter.EnterReferenceNumber,
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

			session := &sesh.LoginSession{IDToken: "id-token", Sub: "supporter-random", Email: "name@example.com"}

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
				SetLogin(r, w, session).
				Return(nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				GetAny(mock.Anything).
				Return(nil, dynamo.NotFoundError{})
			memberStore.EXPECT().
				InvitedMembersByEmail(mock.Anything).
				Return(tc.invites, nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				InfoContext(r.Context(), "login", slog.String("session_id", session.SessionID()))

			err := LoginCallback(logger, client, sessionStore, nil, testNowFn, memberStore)(page.AppData{}, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirect.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestLoginCallbackWhenErrorReturned(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?error=hey&error_description=this%20is%20why&state=my-state", nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "login error", slog.String("error", "hey"), slog.String("error_description", "this is why"))

	err := LoginCallback(logger, nil, nil, nil, testNowFn, nil)(page.AppData{}, w, r)
	assert.Equal(t, errors.New("access denied"), err)
}

func TestLoginCallbackWhenMemberGetAnyErrors(t *testing.T) {
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, nil, testNowFn, memberStore)(page.AppData{}, w, r)

	assert.Equal(t, expectedError, err)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackWhenInvitedMembersByEmailErrors(t *testing.T) {
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, dynamo.NotFoundError{})
	memberStore.EXPECT().
		InvitedMembersByEmail(mock.Anything).
		Return([]*actor.MemberInvite{{}}, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, nil, testNowFn, memberStore)(page.AppData{}, w, r)
	resp := w.Result()

	assert.Error(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackHasMember(t *testing.T) {
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
		GetAny(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
		Return(&actor.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
		Return(nil, dynamo.NotFoundError{})

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.EnterOrganisationName.Format(), resp.Header.Get("Location"))
}

func TestLoginCallbackHasMemberWhenSessionErrors(t *testing.T) {
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, dynamo.NotFoundError{})

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackHasMemberWhenOrganisationGetErrors(t *testing.T) {
	loginSession := &sesh.LoginSession{
		IDToken: "id-token",
		Sub:     "supporter-random",
		Email:   "name@example.com",
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})

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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(ctx).
		Return(&actor.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(ctx).
		Return(nil, expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackHasOrganisation(t *testing.T) {
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
				GetAny(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
				Return(&actor.Member{Email: tc.existingMemberEmail}, nil)

			memberStore.EXPECT().
				Put(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email, OrganisationID: "org-id"}), &actor.Member{Email: tc.loginSessionEmail, LastLoggedInAt: testNow}).
				Return(nil)

			organisationStore := newMockOrganisationStore(t)
			organisationStore.EXPECT().
				Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email})).
				Return(&actor.Organisation{ID: "org-id", Name: "org name"}, nil)

			logger := newMockLogger(t)
			logger.EXPECT().
				InfoContext(mock.Anything, mock.Anything, mock.Anything)

			err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Supporter.Dashboard.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestLoginCallbackHasOrganisationWhenMemberPutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "email@example.com"}, nil)

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
		Return(nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.Member{Email: "email@example.com"}, nil)
	memberStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackHasOrganisationWhenSessionErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		Exchange(r.Context(), "auth-code", "my-nonce").
		Return("id-token", "a JWT", nil)
	client.EXPECT().
		UserInfo(r.Context(), "a JWT").
		Return(onelogin.UserInfo{Sub: "random", Email: "email@example.com"}, nil)

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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.Member{Email: "email@example.com"}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}

func TestLoginCallbackSessionError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=auth-code&state=my-state", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
		Return(nil, expectedError)

	err := LoginCallback(nil, nil, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
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

	err := LoginCallback(nil, client, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
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

	err := LoginCallback(nil, client, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.Member{}, nil)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := LoginCallback(logger, client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}
