package donorpage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Doe", RetrievedAt: now}
	updatedDonor := &donordata.Provided{
		LpaID:                 "lpa-id",
		Donor:                 donordata.Donor{FirstNames: "John", LastName: "Doe"},
		DonorIdentityUserData: userData,
		Tasks:                 donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateInProgress},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), updatedDonor).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(r.Context(), "a-code", "a-nonce").
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(r.Context(), "a-jwt").
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(r.Context(), userInfo).
		Return(userData, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.OneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityNotConfirmed(t *testing.T) {
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	sessionRetrieved := func(t *testing.T) *mockSessionStore {
		sessionStore := newMockSessionStore(t)
		sessionStore.EXPECT().
			OneLogin(mock.Anything).
			Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)
		return sessionStore
	}

	sessionIgnored := func(t *testing.T) *mockSessionStore {
		return nil
	}

	donorStoreIgnored := func(t *testing.T) *mockDonorStore {
		return nil
	}

	testCases := map[string]struct {
		oneLoginClient func(t *testing.T) *mockOneLoginClient
		sessionStore   func(*testing.T) *mockSessionStore
		donorStore     func(*testing.T) *mockDonorStore
		url            string
		error          error
	}{
		"errored on donorStore.Put": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.EXPECT().
					ParseIdentityClaim(mock.Anything, mock.Anything).
					Return(identity.UserData{Status: identity.StatusFailed}, nil)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			donorStore: func(t *testing.T) *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(expectedError)

				return donorStore
			},
			error: expectedError,
		},
		"errored on parse": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(userInfo, nil)
				oneLoginClient.EXPECT().
					ParseIdentityClaim(mock.Anything, mock.Anything).
					Return(identity.UserData{}, expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			error:        expectedError,
			donorStore:   donorStoreIgnored,
		},
		"errored on userinfo": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("id-token", "a-jwt", nil)
				oneLoginClient.EXPECT().
					UserInfo(mock.Anything, mock.Anything).
					Return(onelogin.UserInfo{}, expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			error:        expectedError,
			donorStore:   donorStoreIgnored,
		},
		"errored on exchange": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				oneLoginClient := newMockOneLoginClient(t)
				oneLoginClient.EXPECT().
					Exchange(mock.Anything, mock.Anything, mock.Anything).
					Return("", "", expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			error:        expectedError,
			donorStore:   donorStoreIgnored,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			sessionStore: sessionIgnored,
			donorStore:   donorStoreIgnored,
			error:        errors.New("access denied"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			sessionStore := tc.sessionStore(t)
			oneLoginClient := tc.oneLoginClient(t)

			err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, tc.donorStore(t))(testAppData, w, r, &donordata.Provided{})
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenInsufficientEvidenceReturnCodeClaimPresent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{ReturnCodes: []onelogin.ReturnCodeInfo{{Code: "X"}}}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			Donor:                 donordata.Donor{FirstNames: "John", LastName: "Doe"},
			LpaID:                 "lpa-id",
			DonorIdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			Tasks:                 donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateInProgress},
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(identity.UserData{Status: identity.StatusInsufficientEvidence}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.UnableToConfirmIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenAnyOtherReturnCodeClaimPresent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{ReturnCodes: []onelogin.ReturnCodeInfo{{Code: "T"}}}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			Donor:                 donordata.Donor{FirstNames: "John", LastName: "Doe"},
			LpaID:                 "lpa-id",
			DonorIdentityUserData: identity.UserData{Status: identity.StatusFailed},
			Tasks:                 donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateProblem},
		}).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(identity.UserData{Status: identity.StatusFailed}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.RegisterWithCourtOfProtection.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenPutDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(mock.Anything).
		Return(&sesh.OneLoginSession{State: "a-state", Nonce: "a-nonce", Redirect: "/redirect"}, nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		Exchange(mock.Anything, mock.Anything, mock.Anything).
		Return("id-token", "a-jwt", nil)
	oneLoginClient.EXPECT().
		UserInfo(mock.Anything, mock.Anything).
		Return(userInfo, nil)
	oneLoginClient.EXPECT().
		ParseIdentityClaim(mock.Anything, mock.Anything).
		Return(identity.UserData{Status: identity.StatusConfirmed}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	err := IdentityWithOneLoginCallback(nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:                 "lpa-id",
		Donor:                 donordata.Donor{FirstNames: "first-name", LastName: "last-name"},
		DonorIdentityUserData: userData,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.OneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}
