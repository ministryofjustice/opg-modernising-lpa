package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeAttorneyLoggedOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := lpadata.Lpa{LpaUID: "lpa-uid"}

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})).
		Return(&lpa, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeAttorneyDataLoggedOut{
			App: testAppData,
			Lpa: &lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeAttorneyLoggedOut(template.Execute, nil, lpaStoreResolvingService, sessionStore, nil, "example.com", nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenSessionStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{}, expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, nil, nil, sessionStore, nil, "example.com", nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, nil, lpaStoreResolvingService, sessionStore, nil, "example.com", nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyLoggedOutWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmDontWantToBeAttorneyLoggedOut(template.Execute, nil, lpaStoreResolvingService, sessionStore, nil, "example.com", nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorneyLoggedOut(t *testing.T) {
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	testcases := map[string]struct {
		uid              actoruid.UID
		attorneyFullName string
		actorType        actor.Type
	}{
		"attorney": {
			uid:              attorneyUID,
			attorneyFullName: "d e f",
			actorType:        actor.TypeAttorney,
		},
		"replacement attorney": {
			uid:              replacementAttorneyUID,
			attorneyFullName: "x y z",
			actorType:        actor.TypeReplacementAttorney,
		},
		"trust corporation": {
			uid:              trustCorporationUID,
			attorneyFullName: "trusty",
			actorType:        actor.TypeTrustCorporation,
		},
		"replacement trust corporation": {
			uid:              replacementTrustCorporationUID,
			attorneyFullName: "untrusty",
			actorType:        actor.TypeReplacementTrustCorporation,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
			w := httptest.NewRecorder()
			ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

			lpa := &lpadata.Lpa{
				LpaUID:   "lpa-uid",
				SignedAt: time.Now(),
				Donor: lpadata.Donor{
					FirstNames: "a b", LastName: "c", Email: "a@example.com",
					ContactLanguagePreference: localize.En,
				},
				Attorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{UID: trustCorporationUID, Name: "trusty"},
					Attorneys: []lpadata.Attorney{
						{UID: attorneyUID, FirstNames: "d e", LastName: "f"},
					},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					TrustCorporation: lpadata.TrustCorporation{UID: replacementTrustCorporationUID, Name: "untrusty"},
					Attorneys: []lpadata.Attorney{
						{UID: replacementAttorneyUID, FirstNames: "x y", LastName: "z"},
					},
				},
				Type: lpadata.LpaTypePersonalWelfare,
			}

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				LpaData(r).
				Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

			shareCodeData := sharecodedata.Link{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
				ActorUID:    tc.uid,
			}

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, "123").
				Return(shareCodeData, nil)
			shareCodeStore.EXPECT().
				Delete(r.Context(), shareCodeData).
				Return(nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(ctx).
				Return(lpa, nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(lpa).
				Return("Dear donor")
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.AttorneyOptedOutEmail{
					Greeting:          "Dear donor",
					AttorneyFullName:  tc.attorneyFullName,
					DonorFullName:     "a b c",
					LpaType:           "Personal welfare",
					LpaUID:            "lpa-uid",
					DonorStartPageURL: "example.com" + page.PathStart.Format(),
				}).
				Return(nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare")

			testAppData.Localizer = localizer

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendAttorneyOptOut(r.Context(), "lpa-uid", tc.uid, tc.actorType).
				Return(nil)

			err := ConfirmDontWantToBeAttorneyLoggedOut(nil, shareCodeStore, lpaStoreResolvingService, sessionStore, notifyClient, "example.com", lpaStoreClient)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, page.PathAttorneyYouHaveDecidedNotToBeAttorney.Format()+"?donorFullName=a+b+c", resp.Header.Get("Location"))
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestPostConfirmDontWantToBeAttorneyLoggedOutWhenAttorneyNotFound(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		LpaData(r).
		Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

	shareCodeData := sharecodedata.Link{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")),
		ActorUID:    actoruid.New(),
	}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "123").
		Return(shareCodeData, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(ctx).
		Return(&lpadata.Lpa{
			LpaUID:   "lpa-uid",
			SignedAt: time.Now(),
			Donor: lpadata.Donor{
				FirstNames: "a b", LastName: "c", Email: "a@example.com",
			},
			Type: lpadata.LpaTypePersonalWelfare,
		}, nil)

	err := ConfirmDontWantToBeAttorneyLoggedOut(nil, shareCodeStore, lpaStoreResolvingService, sessionStore, nil, "example.com", nil)(testAppData, w, r)
	assert.EqualError(t, err, "attorney not found")
}

func TestPostConfirmDontWantToBeAttorneyLoggedOutErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)
	ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: "lpa-id"})

	shareCodeData := sharecodedata.Link{
		LpaKey: dynamo.LpaKey("lpa-id"),
	}

	signedLPA := lpadata.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()}

	localizer := func(t *testing.T) *mockLocalizer {
		l := newMockLocalizer(t)
		l.EXPECT().
			T(mock.Anything).
			Return("a")

		return l
	}

	testcases := map[string]struct {
		sessionStore             func(*testing.T) *mockSessionStore
		lpaStoreResolvingService func(*testing.T) *mockLpaStoreResolvingService
		localizer                func(*testing.T) *mockLocalizer
		shareCodeStore           func(*testing.T) *mockShareCodeStore
		notifyClient             func(*testing.T) *mockNotifyClient
		lpaStoreClient           func(*testing.T) *mockLpaStoreClient
	}{
		"when shareCodeStore.Get() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, expectedError)

				return shareCodeStore
			},
		},
		"when notifyClient.SendActorEmail() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			localizer: localizer,
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
		},
		"when lpaStoreClient.SendAttorneyOptOut() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			localizer: localizer,
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)

				return shareCodeStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					SendAttorneyOptOut(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
				return client
			},
		},
		"when shareCodeStore.Delete() error": {
			sessionStore: func(t *testing.T) *mockSessionStore {
				sessionStore := newMockSessionStore(t)
				sessionStore.EXPECT().
					LpaData(r).
					Return(&sesh.LpaDataSession{LpaID: "lpa-id"}, nil)

				return sessionStore
			},
			lpaStoreResolvingService: func(t *testing.T) *mockLpaStoreResolvingService {
				lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
				lpaStoreResolvingService.EXPECT().
					Get(ctx).
					Return(&signedLPA, nil)

				return lpaStoreResolvingService
			},
			localizer: localizer,
			shareCodeStore: func(t *testing.T) *mockShareCodeStore {
				shareCodeStore := newMockShareCodeStore(t)
				shareCodeStore.EXPECT().
					Get(mock.Anything, mock.Anything, mock.Anything).
					Return(shareCodeData, nil)
				shareCodeStore.EXPECT().
					Delete(mock.Anything, mock.Anything).
					Return(expectedError)

				return shareCodeStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("Dear donor")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					SendAttorneyOptOut(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				return client
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			testAppData.Localizer = evalT(tc.localizer, t)

			err := ConfirmDontWantToBeAttorneyLoggedOut(nil, evalT(tc.shareCodeStore, t), evalT(tc.lpaStoreResolvingService, t), evalT(tc.sessionStore, t), evalT(tc.notifyClient, t), "example.com", evalT(tc.lpaStoreClient, t))(testAppData, w, r)

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
