package voucherpage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithOneLoginCallback(t *testing.T) {
	now := time.Now()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Doe", RetrievedAt: now}

	updatedVoucher := &voucherdata.Provided{
		LpaID:            "lpa-id",
		FirstNames:       "John",
		LastName:         "Doe",
		IdentityUserData: userData,
		Tasks:            voucherdata.Tasks{ConfirmYourIdentity: task.StateCompleted},
	}

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), updatedVoucher).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{LpaUID: "lpa-uid", Voucher: lpadata.Voucher{FirstNames: "John", LastName: "Doe"}}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
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

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, voucherStore, lpaStoreResolvingService, nil, "www.example.com")(testAppData, w, r, &voucherdata.Provided{
		LpaID:      "lpa-id",
		FirstNames: "John",
		LastName:   "Doe",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenFailedIdentityCheck(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusFailed}

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{
			LpaUID:   "lpa-uid",
			Voucher:  lpadata.Voucher{FirstNames: "a", LastName: "b"},
			Donor:    lpadata.Donor{Email: "a@example.com", FirstNames: "c", LastName: "d"},
			Type:     lpadata.LpaTypePersonalWelfare,
			SignedAt: time.Now(),
		}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		OneLogin(r).
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

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("translated LPA type")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("Dear donor")
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), "a@example.com", "lpa-uid", notify.VoucherFailedIdentityCheckEmail{
			Greeting:          "Dear donor",
			DonorFullName:     "c d",
			VoucherFullName:   "a b",
			LpaType:           "translated LPA type",
			DonorStartPageURL: "www.example.com" + page.PathStart.Format(),
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, voucherStore, lpaStoreResolvingService, notifyClient, "www.example.com")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, voucher.PathUnableToConfirmIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenSendingEmailError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusFailed}

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{
			LpaUID:   "lpa-uid",
			Voucher:  lpadata.Voucher{FirstNames: "a", LastName: "b"},
			Donor:    lpadata.Donor{Email: "a@example.com", FirstNames: "c", LastName: "d"},
			Type:     lpadata.LpaTypePersonalWelfare,
			SignedAt: time.Now(),
		}, nil)

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
		Return(userData, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("translated LPA type")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("")
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, voucherStore, lpaStoreResolvingService, notifyClient, "www.example.com")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
	voucherIgnored := func(t *testing.T) *mockVoucherStore {
		return nil
	}

	testCases := map[string]struct {
		oneLoginClient      func(t *testing.T) *mockOneLoginClient
		sessionStore        func(*testing.T) *mockSessionStore
		voucherStore        func(t *testing.T) *mockVoucherStore
		url                 string
		error               error
		expectedRedirectURL string
		expectedStatus      int
	}{
		"not ok": {
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
					Return(identity.UserData{}, nil)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			voucherStore: func(t *testing.T) *mockVoucherStore {
				voucherStore := newMockVoucherStore(t)
				voucherStore.EXPECT().
					Put(context.Background(), &voucherdata.Provided{
						LpaID: "lpa-id",
						Tasks: voucherdata.Tasks{ConfirmYourIdentity: task.StateCompleted},
					}).
					Return(nil)

				return voucherStore
			},
			expectedRedirectURL: voucher.PathUnableToConfirmIdentity.Format("lpa-id"),
			expectedStatus:      http.StatusFound,
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
					Return(identity.UserData{Status: identity.StatusConfirmed}, expectedError)
				return oneLoginClient
			},
			sessionStore:   sessionRetrieved,
			error:          expectedError,
			voucherStore:   voucherIgnored,
			expectedStatus: http.StatusOK,
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
			sessionStore:   sessionRetrieved,
			error:          expectedError,
			voucherStore:   voucherIgnored,
			expectedStatus: http.StatusOK,
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
			sessionStore:   sessionRetrieved,
			error:          expectedError,
			voucherStore:   voucherIgnored,
			expectedStatus: http.StatusOK,
		},
		"errored on session store": {
			url: "/?code=a-code",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return nil
			},
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					OneLogin(mock.Anything).
					Return(nil, expectedError)
				return sessionStore
			},
			error:          expectedError,
			voucherStore:   voucherIgnored,
			expectedStatus: http.StatusOK,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			sessionStore:   sessionIgnored,
			error:          errors.New("access denied"),
			voucherStore:   voucherIgnored,
			expectedStatus: http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(&lpadata.Lpa{Voucher: lpadata.Voucher{}}, nil)

			sessionStore := tc.sessionStore(t)
			oneLoginClient := tc.oneLoginClient(t)

			err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, tc.voucherStore(t), lpaStoreResolvingService, nil, "www.example.com")(testAppData, w, r, &voucherdata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirectURL, resp.Header.Get("Location"))
		})
	}
}

func TestGetIdentityWithOneLoginCallbackWhenGetLpaStoreResolvingServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Voucher: lpadata.Voucher{}}, expectedError)

	err := IdentityWithOneLoginCallback(nil, nil, nil, lpaStoreResolvingService, nil, "www.example.com")(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenPutVoucherStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}

	voucherStore := newMockVoucherStore(t)
	voucherStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpadata.Lpa{Voucher: lpadata.Voucher{}}, nil)

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

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, voucherStore, lpaStoreResolvingService, nil, "www.example.com")(testAppData, w, r, &voucherdata.Provided{})

	assert.Equal(t, expectedError, err)
}
