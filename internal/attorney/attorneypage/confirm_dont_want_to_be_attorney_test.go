package attorneypage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmDontWantToBeAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{LpaUID: "lpa-uid"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmDontWantToBeAttorneyData{
			App: testAppData,
			Lpa: lpa,
		}).
		Return(nil)

	err := ConfirmDontWantToBeAttorney(template.Execute, nil, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmDontWantToBeAttorneyWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ConfirmDontWantToBeAttorney(template.Execute, nil, nil, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorney(t *testing.T) {
	r, _ := http.NewRequestWithContext(appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()

	uid := actoruid.New()
	lpa := &lpadata.Lpa{
		LpaUID:   "lpa-uid",
		SignedAt: time.Now(),
		Donor: lpadata.Donor{
			FirstNames: "a b", LastName: "c", Email: "a@example.com",
			ContactLanguagePreference: localize.En,
		},
		Attorneys: lpadata.Attorneys{
			Attorneys: []lpadata.Attorney{
				{FirstNames: "d e", LastName: "f", UID: uid},
			},
		},
		Type: lpadata.LpaTypePersonalWelfare,
	}

	certificateProviderStore := newMockAttorneyStore(t)
	certificateProviderStore.EXPECT().
		Delete(r.Context()).
		Return(nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare")

	testAppData.Localizer = localizer

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("Dear donor")
	notifyClient.EXPECT().
		SendActorEmail(r.Context(), notify.ToLpaDonor(lpa), "lpa-uid", notify.AttorneyOptedOutEmail{
			Greeting:           "Dear donor",
			AttorneyFullName:   "d e f",
			LpaType:            "Personal welfare",
			LpaReferenceNumber: "lpa-uid",
		}).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendAttorneyOptOut(r.Context(), "lpa-uid", uid, actor.TypeAttorney).
		Return(nil)

	err := ConfirmDontWantToBeAttorney(nil, certificateProviderStore, notifyClient, lpaStoreClient)(testAppData, w, r, &attorneydata.Provided{
		UID: uid,
	}, lpa)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, page.PathAttorneyYouHaveDecidedNotToBeAttorney.Format()+"?donorFirstNames=a+b&donorFullName=a+b+c", resp.Header.Get("Location"))
	assert.Equal(t, http.StatusFound, resp.StatusCode)
}

func TestPostConfirmDontWantToBeAttorneyWhenAttorneyNotFound(t *testing.T) {
	r, _ := http.NewRequestWithContext(appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123", SessionID: "456"}), http.MethodPost, "/?referenceNumber=123", nil)
	w := httptest.NewRecorder()
	uid := actoruid.New()

	err := ConfirmDontWantToBeAttorney(nil, nil, nil, nil)(testAppData, w, r, &attorneydata.Provided{
		UID: uid,
	}, &lpadata.Lpa{
		LpaUID:   "lpa-uid",
		SignedAt: time.Now(),
		Donor: lpadata.Donor{
			FirstNames: "a b", LastName: "c", Email: "a@example.com",
		},
		Type: lpadata.LpaTypePersonalWelfare,
	})
	assert.EqualError(t, err, "attorney not found")
}

func TestPostConfirmDontWantToBeAttorneyErrors(t *testing.T) {
	r, _ := http.NewRequest(http.MethodPost, "/?referenceNumber=123", nil)

	testcases := map[string]struct {
		attorneyStore  func(*testing.T) *mockAttorneyStore
		notifyClient   func(*testing.T) *mockNotifyClient
		lpaStoreClient func(*testing.T) *mockLpaStoreClient
	}{
		"when notifyClient.SendActorEmail() error": {
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
		},
		"when lpaStoreClient.SendAttorneyOptOut() error": {
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					SendAttorneyOptOut(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)
				return client
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
		},
		"when attorneyStore.Delete() error": {
			lpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				client := newMockLpaStoreClient(t)
				client.EXPECT().
					SendAttorneyOptOut(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				return client
			},
			attorneyStore: func(t *testing.T) *mockAttorneyStore {
				attorneyStore := newMockAttorneyStore(t)
				attorneyStore.EXPECT().
					Delete(mock.Anything).
					Return(expectedError)

				return attorneyStore
			},
			notifyClient: func(t *testing.T) *mockNotifyClient {
				client := newMockNotifyClient(t)
				client.EXPECT().
					EmailGreeting(mock.Anything).
					Return("")
				client.EXPECT().
					SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return client
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(mock.Anything).
				Return("a")

			testAppData.Localizer = localizer

			err := ConfirmDontWantToBeAttorney(nil, evalT(tc.attorneyStore, t), evalT(tc.notifyClient, t), evalT(tc.lpaStoreClient, t))(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{LpaUID: "lpa-uid", SignedAt: time.Now()})

			resp := w.Result()

			assert.Equal(t, expectedError, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
