package donor

import (
	io "io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "John Doe", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			OneLoginUserData: userData,
		}).
		Return(nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", r.Context(), "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", r.Context(), userInfo).
		Return(userData, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	testCases := map[string]struct {
		oneLoginClient func(t *testing.T) *mockOneLoginClient
		template       func(*testing.T, io.Writer) *mockTemplate
		url            string
		error          error
	}{
		"not ok": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{}, nil)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				template := newMockTemplate(t)
				template.
					On("Execute", w, &identityWithOneLoginCallbackData{
						App:             testAppData,
						CouldNotConfirm: true,
					}).
					Return(nil)
				return template
			},
		},
		"errored on parse": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.
					On("ParseIdentityClaim", mock.Anything, mock.Anything).
					Return(identity.UserData{OK: true}, expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},

			error: expectedError,
		},
		"errored on userinfo": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("a-jwt", nil)
				oneLoginClient.
					On("UserInfo", mock.Anything, mock.Anything).
					Return(onelogin.UserInfo{}, expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},

			error: expectedError,
		},
		"errored on exchange": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.
					On("Exchange", mock.Anything, mock.Anything, mock.Anything).
					Return("", expectedError)
				return oneLoginClient
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				return nil
			},

			error: expectedError,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			template: func(t *testing.T, w io.Writer) *mockTemplate {
				template := newMockTemplate(t)
				template.
					On("Execute", w, &identityWithOneLoginCallbackData{
						App:             testAppData,
						CouldNotConfirm: true,
					}).
					Return(nil)
				return template
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)

			sessionStore := &mockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
					},
				}, nil)

			oneLoginClient := tc.oneLoginClient(t)
			template := tc.template(t, w)

			err := IdentityWithOneLoginCallback(template.Execute, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{OneLoginUserData: userData}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithOneLoginCallbackData{
			App:         testAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Execute, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{
		OneLoginUserData: identity.UserData{OK: true},
	}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.SelectYourIdentityOptions1, resp.Header.Get("Location"))
}
