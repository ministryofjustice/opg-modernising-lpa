package donorpage

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
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
		PK:               dynamo.LpaKey("hey"),
		SK:               dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID:            "lpa-id",
		Donor:            donordata.Donor{FirstNames: "John", LastName: "Doe"},
		IdentityUserData: userData,
		Tasks:            donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateInProgress},
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
		ParseIdentityClaim(userInfo).
		Return(userData, nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Put(r.Context(), scheduled.Event{
			//At:                now.AddDate(0, 6, 0),
			At:                now,
			Action:            scheduled.ActionExpireDonorIdentity,
			TargetLpaKey:      dynamo.LpaKey("hey"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, scheduledStore, nil)(testAppData, w, r, &donordata.Provided{
		PK:    dynamo.LpaKey("hey"),
		SK:    dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityMismatched(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	actorUID := actoruid.New()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Does", RetrievedAt: now}
	updatedDonor := &donordata.Provided{
		PK:                               dynamo.LpaKey("hey"),
		SK:                               dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID:                            "lpa-id",
		LpaUID:                           "lpa-uid",
		Donor:                            donordata.Donor{UID: actorUID, FirstNames: "John", LastName: "Doe"},
		IdentityUserData:                 userData,
		Tasks:                            donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateInProgress},
		WitnessedByCertificateProviderAt: testNow,
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
		ParseIdentityClaim(userInfo).
		Return(userData, nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Put(r.Context(), scheduled.Event{
			//At:                now.AddDate(0, 6, 0),
			At:                now,
			Action:            scheduled.ActionExpireDonorIdentity,
			TargetLpaKey:      dynamo.LpaKey("hey"),
			TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendIdentityCheckMismatched(r.Context(), event.IdentityCheckMismatched{
			LpaUID:   "lpa-uid",
			ActorUID: actoruid.Prefixed(actorUID),
			Provided: event.IdentityCheckMismatchedDetails{
				FirstNames: "John",
				LastName:   "Doe",
			},
			Verified: event.IdentityCheckMismatchedDetails{
				FirstNames: "John",
				LastName:   "Does",
			},
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, scheduledStore, eventClient)(testAppData, w, r, &donordata.Provided{
		PK:                               dynamo.LpaKey("hey"),
		SK:                               dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID:                            "lpa-id",
		LpaUID:                           "lpa-uid",
		Donor:                            donordata.Donor{UID: actorUID, FirstNames: "John", LastName: "Doe"},
		WitnessedByCertificateProviderAt: testNow,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenIdentityMismatchedEventErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	actorUID := actoruid.New()
	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Does", RetrievedAt: now}

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
		ParseIdentityClaim(userInfo).
		Return(userData, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendIdentityCheckMismatched(mock.Anything, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, nil, nil, eventClient)(testAppData, w, r, &donordata.Provided{
		PK:                               dynamo.LpaKey("hey"),
		SK:                               dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID:                            "lpa-id",
		LpaUID:                           "lpa-uid",
		Donor:                            donordata.Donor{UID: actorUID, FirstNames: "John", LastName: "Doe"},
		WitnessedByCertificateProviderAt: testNow,
	})
	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenScheduledStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Now()

	userInfo := onelogin.UserInfo{CoreIdentityJWT: "an-identity-jwt"}
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "John", LastName: "Doe", RetrievedAt: now}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
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
		ParseIdentityClaim(mock.Anything).
		Return(userData, nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, scheduledStore, nil)(testAppData, w, r, &donordata.Provided{
		PK:    dynamo.LpaKey("hey"),
		SK:    dynamo.LpaOwnerKey(dynamo.DonorKey("oh")),
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
	})

	assert.Equal(t, expectedError, err)
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

	sessionIgnored := func(*testing.T) *mockSessionStore { return nil }
	donorStoreIgnored := func(*testing.T) *mockDonorStore { return nil }
	eventClientIgnored := func(*testing.T) *mockEventClient { return nil }

	testCases := map[string]struct {
		oneLoginClient func(t *testing.T) *mockOneLoginClient
		sessionStore   func(*testing.T) *mockSessionStore
		donorStore     func(*testing.T) *mockDonorStore
		eventClient    func(*testing.T) *mockEventClient
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
					ParseIdentityClaim(mock.Anything).
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
			eventClient: func(t *testing.T) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendIdentityCheckMismatched(mock.Anything, mock.Anything).
					Return(nil)

				return eventClient
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
					ParseIdentityClaim(mock.Anything).
					Return(identity.UserData{}, expectedError)
				return oneLoginClient
			},
			sessionStore: sessionRetrieved,
			error:        expectedError,
			eventClient:  eventClientIgnored,
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
			eventClient:  eventClientIgnored,
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
			eventClient:  eventClientIgnored,
			donorStore:   donorStoreIgnored,
		},
		"provider access denied": {
			url: "/?error=access_denied",
			oneLoginClient: func(t *testing.T) *mockOneLoginClient {
				return newMockOneLoginClient(t)
			},
			sessionStore: sessionIgnored,
			eventClient:  eventClientIgnored,
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
			eventClient := tc.eventClient(t)

			err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, tc.donorStore(t), nil, eventClient)(testAppData, w, r, &donordata.Provided{})
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
			Donor:            donordata.Donor{FirstNames: "John", LastName: "Doe"},
			LpaID:            "lpa-id",
			IdentityUserData: identity.UserData{Status: identity.StatusInsufficientEvidence},
			Tasks:            donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateInProgress},
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
		ParseIdentityClaim(mock.Anything).
		Return(identity.UserData{Status: identity.StatusInsufficientEvidence}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, nil, nil)(testAppData, w, r, &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "John", LastName: "Doe"},
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathUnableToConfirmIdentity.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetIdentityWithOneLoginCallbackWhenAnyOtherReturnCodeClaimPresent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	userInfo := onelogin.UserInfo{ReturnCodes: []onelogin.ReturnCodeInfo{{Code: "T"}}}
	actorUID := actoruid.New()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			Donor:            donordata.Donor{UID: actorUID, FirstNames: "John", LastName: "Doe"},
			LpaID:            "lpa-id",
			LpaUID:           "lpa-uid",
			IdentityUserData: identity.UserData{Status: identity.StatusFailed},
			Tasks:            donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateProblem},
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
		ParseIdentityClaim(mock.Anything).
		Return(identity.UserData{Status: identity.StatusFailed}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendIdentityCheckMismatched(r.Context(), event.IdentityCheckMismatched{
			LpaUID:   "lpa-uid",
			ActorUID: actoruid.Prefixed(actorUID),
			Provided: event.IdentityCheckMismatchedDetails{
				FirstNames: "John",
				LastName:   "Doe",
			},
		}).
		Return(nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, nil, eventClient)(testAppData, w, r, &donordata.Provided{
		Donor:  donordata.Donor{UID: actorUID, FirstNames: "John", LastName: "Doe"},
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathRegisterWithCourtOfProtection.Format("lpa-id"), resp.Header.Get("Location"))
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
		ParseIdentityClaim(mock.Anything).
		Return(identity.UserData{Status: identity.StatusConfirmed}, nil)

	err := IdentityWithOneLoginCallback(oneLoginClient, sessionStore, donorStore, nil, nil)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithOneLoginCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?code=a-code", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{Status: identity.StatusConfirmed, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	err := IdentityWithOneLoginCallback(nil, nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:            "lpa-id",
		Donor:            donordata.Donor{FirstNames: "first-name", LastName: "last-name"},
		IdentityUserData: userData,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathOneLoginIdentityDetails.Format("lpa-id"), resp.Header.Get("Location"))
}
