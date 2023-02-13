package donor

import (
	"context"
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

type mockOneLoginClient struct {
	mock.Mock
}

func (m *mockOneLoginClient) AuthCodeURL(state, nonce, locale string, identity bool) string {
	args := m.Called(state, nonce, locale, identity)
	return args.String(0)
}

func (m *mockOneLoginClient) Exchange(ctx context.Context, code, nonce string) (string, error) {
	args := m.Called(ctx, code, nonce)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockOneLoginClient) UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(onelogin.UserInfo), args.Error(1)
}

func (m *mockOneLoginClient) ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error) {
	args := m.Called(ctx, userInfo)
	return args.Get(0).(identity.UserData), args.Error(1)
}

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{OK: true, FullName: "John Doe", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
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

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", r.Context(), "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", r.Context(), userInfo).
		Return(userData, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithOneLoginCallbackData{
			App:         appData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient, template)
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	testCases := map[string]struct {
		userData identity.UserData
		url      string
		error    error
	}{
		"not ok": {
			url: "/?code=a-code",
		},
		"errored": {
			url:      "/?code=a-code",
			userData: identity.UserData{OK: true},
			error:    expectedError,
		},
		"provider access denied": {
			url:      "/?error=access_denied",
			userData: identity.UserData{OK: true},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

			lpaStore := &mockLpaStore{}
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

			oneLoginClient := &mockOneLoginClient{}
			oneLoginClient.
				On("Exchange", mock.Anything, mock.Anything, mock.Anything).
				Return("a-jwt", nil)
			oneLoginClient.
				On("UserInfo", mock.Anything, mock.Anything).
				Return(userInfo, nil)
			oneLoginClient.
				On("ParseIdentityClaim", mock.Anything, mock.Anything).
				Return(tc.userData, tc.error)

			template := &mockTemplate{}
			template.
				On("Func", w, &identityWithOneLoginCallbackData{
					App:             appData,
					CouldNotConfirm: true,
				}).
				Return(nil)

			err := IdentityWithOneLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenExchangeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
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

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("", expectedError)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
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

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{}, expectedError)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithOneLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &mockLpaStore{}
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

	oneLoginClient := &mockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{OneLoginUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithOneLoginCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Func, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{
		OneLoginUserData: identity.UserData{OK: true},
	}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.SelectYourIdentityOptions1, resp.Header.Get("Location"))
}
