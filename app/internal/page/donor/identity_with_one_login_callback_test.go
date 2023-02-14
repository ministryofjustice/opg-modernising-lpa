package donor

import (
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

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			OneLoginUserData: userData,
		}).
		Return(nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := &page.MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", r.Context(), "a-code", "a-nonce").
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", r.Context(), userInfo).
		Return(userData, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &identityWithOneLoginCallbackData{
			App:         page.TestAppData,
			FullName:    "John Doe",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(page.TestAppData, w, r)
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
			error:    page.ExpectedError,
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

			lpaStore := &page.MockLpaStore{}
			lpaStore.
				On("Get", r.Context()).
				Return(&page.Lpa{}, nil)

			sessionStore := &page.MockSessionsStore{}
			sessionStore.
				On("Get", mock.Anything, "params").
				Return(&sessions.Session{
					Values: map[any]any{
						"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
					},
				}, nil)

			oneLoginClient := &page.MockOneLoginClient{}
			oneLoginClient.
				On("Exchange", mock.Anything, mock.Anything, mock.Anything).
				Return("a-jwt", nil)
			oneLoginClient.
				On("UserInfo", mock.Anything, mock.Anything).
				Return(userInfo, nil)
			oneLoginClient.
				On("ParseIdentityClaim", mock.Anything, mock.Anything).
				Return(tc.userData, tc.error)

			template := &page.MockTemplate{}
			template.
				On("Func", w, &identityWithOneLoginCallbackData{
					App:             page.TestAppData,
					CouldNotConfirm: true,
				}).
				Return(nil)

			err := IdentityWithOneLoginCallback(template.Func, oneLoginClient, sessionStore, lpaStore)(page.TestAppData, w, r)
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenExchangeError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := &page.MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("", page.ExpectedError)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenUserInfoError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := &page.MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(onelogin.UserInfo{}, page.ExpectedError)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, page.ExpectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithOneLoginCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), mock.Anything).
		Return(page.ExpectedError)

	sessionStore := &page.MockSessionsStore{}
	sessionStore.
		On("Get", mock.Anything, "params").
		Return(&sessions.Session{
			Values: map[any]any{
				"one-login": &sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce"},
			},
		}, nil)

	oneLoginClient := &page.MockOneLoginClient{}
	oneLoginClient.
		On("Exchange", mock.Anything, mock.Anything, mock.Anything).
		Return("a-jwt", nil)
	oneLoginClient.
		On("UserInfo", mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.
		On("ParseIdentityClaim", mock.Anything, mock.Anything).
		Return(identity.UserData{OK: true}, nil)

	err := IdentityWithOneLoginCallback(nil, oneLoginClient, sessionStore, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, oneLoginClient)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{OneLoginUserData: userData}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &identityWithOneLoginCallbackData{
			App:         page.TestAppData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(template.Func, nil, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{
		OneLoginUserData: identity.UserData{OK: true},
	}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}

func TestPostIdentityWithOneLoginCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.SelectYourIdentityOptions1, resp.Header.Get("Location"))
}
