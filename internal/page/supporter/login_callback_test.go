package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
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
		getError      error
		redirect      string
		expectedError error
		loginSession  *sesh.LoginSession
	}{
		"no organisation": {
			getError: dynamo.NotFoundError{},
			redirect: page.Paths.Supporter.EnterOrganisationName.Format(),
			loginSession: &sesh.LoginSession{
				IDToken: "id-token",
				Sub:     "supporter-random",
				Email:   "name@example.com",
			},
		},
		"error getting organisation": {
			getError:      expectedError,
			expectedError: expectedError,
			loginSession: &sesh.LoginSession{
				IDToken: "id-token",
				Sub:     "supporter-random",
				Email:   "name@example.com",
			},
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
				Return(onelogin.UserInfo{Sub: "random", Email: "name@example.com"}, nil)

			sessionStore := newMockSessionStore(t)

			session := sessions.NewSession(sessionStore, "session")
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   true,
			}
			session.Values = map[any]any{"session": tc.loginSession}

			sessionStore.EXPECT().
				Get(r, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{
							State:    "my-state",
							Nonce:    "my-nonce",
							Locale:   "en",
							Redirect: "/redirect",
						},
					},
				}, nil)

			if tc.expectedError == nil {
				sessionStore.EXPECT().
					Save(r, w, session).
					Return(nil)
			}

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				InvitedMember(mock.Anything).
				Return(nil, expectedError)

			organisationStore := newMockOrganisationStore(t)
			organisationStore.EXPECT().
				Get(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: tc.loginSession.SessionID(), Email: tc.loginSession.Email})).
				Return(&actor.Organisation{ID: "org-id", Name: "org name"}, tc.getError)

			err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err)
			} else {
				assert.Nil(t, err)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, tc.redirect, resp.Header.Get("Location"))
			}
		})
	}
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
		Get(r, mock.Anything).
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				},
			},
		}, nil)

	sessionStore.EXPECT().
		Save(r, w, mock.Anything).
		Return(expectedError)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(nil, expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{ID: "org-id", Name: "org name"}, dynamo.NotFoundError{})

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
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

			sessionStore := newMockSessionStore(t)

			session := sessions.NewSession(sessionStore, "session")
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   true,
			}

			loginSession := &sesh.LoginSession{
				IDToken:          "id-token",
				Sub:              "supporter-random",
				Email:            tc.loginSessionEmail,
				OrganisationID:   "org-id",
				OrganisationName: "org name",
			}

			session.Values = map[any]any{"session": loginSession}

			sessionStore.EXPECT().
				Get(r, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{
							State:    "my-state",
							Nonce:    "my-nonce",
							Locale:   "en",
							Redirect: "/redirect",
						},
					},
				}, nil)

			sessionStore.EXPECT().
				Save(r, w, session).
				Return(nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				InvitedMember(mock.Anything).
				Return(nil, expectedError)

			memberStore.EXPECT().
				Self(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email, OrganisationID: "org-id"})).
				Return(&actor.Member{Email: tc.existingMemberEmail}, nil)

			memberStore.EXPECT().
				PutMember(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email, OrganisationID: "org-id"}), &actor.Member{Email: tc.loginSessionEmail, LastLoggedInAt: testNow}).
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
		"Self error": {
			memberError: expectedError,
		},
		"PutMember error": {
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

			sessionStore := newMockSessionStore(t)

			session := sessions.NewSession(sessionStore, "session")
			session.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
				Secure:   true,
			}

			loginSession := &sesh.LoginSession{
				IDToken:          "id-token",
				Sub:              "supporter-random",
				Email:            "name@example.com",
				OrganisationID:   "org-id",
				OrganisationName: "org name",
			}

			session.Values = map[any]any{"session": loginSession}

			sessionStore.EXPECT().
				Get(r, mock.Anything).
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{
							State:    "my-state",
							Nonce:    "my-nonce",
							Locale:   "en",
							Redirect: "/redirect",
						},
					},
				}, nil)

			sessionStore.EXPECT().
				Save(mock.Anything, mock.Anything, mock.Anything).
				Return(nil)

			memberStore := newMockMemberStore(t)
			memberStore.EXPECT().
				InvitedMember(mock.Anything).
				Return(nil, expectedError)

			memberStore.EXPECT().
				Self(mock.Anything).
				Return(&actor.Member{Email: loginSession.Email}, tc.memberError)

			if tc.putMemberError != nil {
				memberStore.EXPECT().
					PutMember(mock.Anything, mock.Anything).
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

	session := sessions.NewSession(sessionStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"session": &sesh.LoginSession{
		IDToken: "id-token",
		Sub:     "supporter-random",
		Email:   "name@example.com",
	}}

	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				},
			},
		}, nil)

	sessionStore.EXPECT().
		Save(r, w, session).
		Return(nil)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(nil, nil)

	err := LoginCallback(client, sessionStore, nil, testNowFn, memberStore)(page.AppData{}, w, r)

	assert.Nil(t, err)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.EnterReferenceNumber.Format(), resp.Header.Get("Location"))
}

func TestLoginCallbackWhenEmailHasInviteWhenSetLoginSessionError(t *testing.T) {
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

	session := sessions.NewSession(sessionStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"session": &sesh.LoginSession{
		IDToken: "id-token",
		Sub:     "supporter-random",
		Email:   "name@example.com",
	}}

	sessionStore.EXPECT().
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: "/redirect",
				},
			},
		}, nil)

	sessionStore.EXPECT().
		Save(r, w, session).
		Return(expectedError)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(nil, nil)

	err := LoginCallback(client, sessionStore, nil, testNowFn, memberStore)(page.AppData{}, w, r)

	assert.Equal(t, expectedError, err)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginCallbackSessionMissing(t *testing.T) {
	testCases := map[string]struct {
		url         string
		session     *sessions.Session
		getErr      error
		expectedErr interface{}
	}{
		"missing session": {
			url:         "/?code=auth-code&state=my-state",
			session:     nil,
			getErr:      expectedError,
			expectedErr: expectedError,
		},
		"missing state": {
			url:         "/?code=auth-code&state=my-state",
			session:     &sessions.Session{Values: map[any]any{}},
			expectedErr: sesh.MissingSessionError("one-login"),
		},
		"missing state from url": {
			url: "/?code=auth-code",
			session: &sessions.Session{
				Values: map[any]any{
					"one-login": &sesh.OneLoginSession{State: "my-state"},
				},
			},
			expectedErr: sesh.InvalidSessionError("one-login"),
		},
		"missing nonce": {
			url: "/?code=auth-code&state=my-state",
			session: &sessions.Session{
				Values: map[any]any{
					"one-login": &sesh.OneLoginSession{State: "my-state", Locale: "en"},
				},
			},
			expectedErr: sesh.InvalidSessionError("one-login"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Get(r, "params").
				Return(tc.session, tc.getErr)

			err := LoginCallback(nil, sessionStore, nil, testNowFn, nil)(page.AppData{}, w, r)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
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
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: page.Paths.Supporter.LoginCallback.Format()},
			},
		}, nil)

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
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "my-state", Nonce: "my-nonce", Locale: "en", Redirect: page.Paths.Supporter.LoginCallback.Format()},
			},
		}, nil)

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
		Get(r, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{
					State:    "my-state",
					Nonce:    "my-nonce",
					Locale:   "en",
					Redirect: page.Paths.Supporter.LoginCallback.Format(),
				},
			},
		}, nil)
	sessionStore.EXPECT().
		Save(r, w, mock.Anything).
		Return(expectedError)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(nil, expectedError)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		Get(mock.Anything).
		Return(&actor.Organisation{}, nil)

	err := LoginCallback(client, sessionStore, organisationStore, testNowFn, memberStore)(page.AppData{}, w, r)
	assert.Equal(t, expectedError, err)
}
