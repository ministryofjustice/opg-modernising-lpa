package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnterAccessCode(t *testing.T) {
	testcases := map[string]struct {
		accessCode         accesscodedata.Link
		session            *sesh.LoginSession
		isReplacement      bool
		isTrustCorporation bool
	}{
		"attorney": {
			accessCode: accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, LpaUID: "lpa-uid"},
			session:    &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
		},
		"replacement": {
			accessCode:    accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, LpaUID: "lpa-uid"},
			session:       &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement: true,
		},
		"trust corporation": {
			accessCode:         accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			accessCode:         accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			session := &sesh.LoginSession{Sub: "hey", Email: "a@example.com"}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Create(r.Context(), tc.accessCode, "a@example.com").
				Return(&attorneydata.Provided{}, nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineAttorney).
				Return(nil)

			err := EnterAccessCode(attorneyStore, nil, eventClient)(testAppData, w, r, session, &lpadata.Lpa{}, tc.accessCode)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathCodeOfConduct.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestEnterAccessCodeWhenAttorneyAlreadySubmittedOnPaper(t *testing.T) {
	testcases := map[string]struct {
		accessCode         accesscodedata.Link
		session            *sesh.LoginSession
		isReplacement      bool
		isTrustCorporation bool
	}{
		"attorney": {
			accessCode: accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, LpaUID: "lpa-uid"},
			session:    &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
		},
		"replacement": {
			accessCode:    accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, LpaUID: "lpa-uid"},
			session:       &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement: true,
		},
		"trust corporation": {
			accessCode:         accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			accessCode:         accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpa := &lpadata.Lpa{Attorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{UID: testUID, Channel: lpadata.ChannelPaper, SignedAt: &testNow}},
			}}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendPaperAttorneyAccessOnline(r.Context(), "lpa-uid", "a@example.com", testUID).
				Return(nil)

			err := EnterAccessCode(nil, lpaStoreClient, nil)(testAppData, w, r, tc.session, lpa, tc.accessCode)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.PathDashboard.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestEnterAccessCodeOnLpaStoreSendPaperAttorneyAccessOnlineError(t *testing.T) {
	accessCode := accesscodedata.Link{ActorUID: testUID}
	lpa := &lpadata.Lpa{Attorneys: lpadata.Attorneys{
		Attorneys: []lpadata.Attorney{{UID: testUID, Channel: lpadata.ChannelPaper, SignedAt: &testNow}}},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendPaperAttorneyAccessOnline(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, lpaStoreClient, nil)(testAppData, w, r, &sesh.LoginSession{}, lpa, accessCode)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnterAccessCodeOnAttorneyStoreError(t *testing.T) {
	accessCode := accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&attorneydata.Provided{}, expectedError)

	err := EnterAccessCode(attorneyStore, nil, nil)(testAppData, w, r, &sesh.LoginSession{}, &lpadata.Lpa{}, accessCode)
	assert.Equal(t, expectedError, err)
}

func TestEnterAccessCodeOnEventClientError(t *testing.T) {
	accessCode := accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&attorneydata.Provided{}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(attorneyStore, nil, eventClient)(testAppData, w, r, &sesh.LoginSession{}, &lpadata.Lpa{}, accessCode)
	assert.ErrorIs(t, err, expectedError)
}
