package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnterAccessCode(t *testing.T) {
	uid := actoruid.New()
	shareCode := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
	session := &sesh.LoginSession{Sub: "hey", Email: "a@example.com"}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Create(r.Context(), shareCode, "a@example.com").
		Return(&certificateproviderdata.Provided{}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineCertificateProvider).
		Return(nil)

	err := EnterAccessCode(nil, certificateProviderStore, nil, nil, eventClient)(testAppData, w, r, session, &lpadata.Lpa{}, shareCode)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathWhoIsEligible.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestEnterAccessCodeWhenPaperCertificateExists(t *testing.T) {
	testcases := map[string]struct {
		donorResults []dashboarddata.Actor
		redirectURL  string
	}{
		"with LPAs": {
			donorResults: []dashboarddata.Actor{{}},
			redirectURL:  page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn.Format() + "?donorFullName=a+b&lpaType=property-and-affairs",
		},
		"without LPAs": {
			redirectURL: page.PathCertificateProviderYouHaveAlreadyProvidedACertificate.Format() + "?donorFullName=a+b&lpaType=property-and-affairs",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			uid := actoruid.New()
			shareCode := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
			session := &sesh.LoginSession{Email: "a@example.com"}

			lpa := &lpadata.Lpa{
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper, SignedAt: &testNow},
				Type:                lpadata.LpaTypePropertyAndAffairs,
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", nil)

			dashboardStore := newMockDashboardStore(t)
			dashboardStore.EXPECT().
				GetAll(r.Context()).
				Return(dashboarddata.Results{Donor: tc.donorResults}, nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				SendPaperCertificateProviderAccessOnline(r.Context(), lpa, "a@example.com").
				Return(nil)

			sessionStore := newMockSessionStore(t)
			if len(tc.donorResults) == 0 {
				sessionStore.EXPECT().
					ClearLogin(r, w).
					Return(nil)
			}

			err := EnterAccessCode(sessionStore, nil, lpaStoreClient, dashboardStore, nil)(testAppData, w, r, session, lpa, shareCode)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirectURL, resp.Header.Get("Location"))
		})
	}
}

func TestEnterAccessCodeWhenSendPaperCertificateProviderAccessOnlineError(t *testing.T) {
	uid := actoruid.New()
	shareCode := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
	lpa := &lpadata.Lpa{
		Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
		CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper, SignedAt: &testNow},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	}
	session := &sesh.LoginSession{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, nil, lpaStoreClient, nil, nil)(testAppData, w, r, session, lpa, shareCode)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnterAccessCodeWhenDashboardStoreError(t *testing.T) {
	uid := actoruid.New()
	shareCode := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
	lpa := &lpadata.Lpa{
		Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
		CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper, SignedAt: &testNow},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	}
	session := &sesh.LoginSession{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{}, expectedError)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := EnterAccessCode(nil, nil, lpaStoreClient, dashboardStore, nil)(testAppData, w, r, session, lpa, shareCode)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestEnterAccessCodeWhenClearLoginError(t *testing.T) {
	uid := actoruid.New()
	shareCode := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
	lpa := &lpadata.Lpa{
		Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
		CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper, SignedAt: &testNow},
		Type:                lpadata.LpaTypePropertyAndAffairs,
	}
	session := &sesh.LoginSession{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	dashboardStore := newMockDashboardStore(t)
	dashboardStore.EXPECT().
		GetAll(r.Context()).
		Return(dashboarddata.Results{}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		ClearLogin(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(sessionStore, nil, lpaStoreClient, dashboardStore, nil)(testAppData, w, r, session, lpa, shareCode)
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeWhenCreateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterAccessCode(nil, certificateProviderStore, nil, nil, nil)(testAppData, w, r, &sesh.LoginSession{}, &lpadata.Lpa{}, sharecodedata.Link{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterAccessCodeWhenEventClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, certificateProviderStore, nil, nil, eventClient)(testAppData, w, r, &sesh.LoginSession{}, &lpadata.Lpa{}, sharecodedata.Link{})
	assert.ErrorIs(t, err, expectedError)
}
