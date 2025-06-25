package attorneypage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockSessionStore) ExpectGet(r *http.Request, values *sesh.LoginSession, err error) {
	m.EXPECT().
		Login(r).
		Return(values, err)
}

func TestGetEnterReferenceNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  testAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReferenceNumberOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  testAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(expectedError)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	testcases := map[string]struct {
		shareCode          sharecodedata.Link
		session            *sesh.LoginSession
		isReplacement      bool
		isTrustCorporation bool
	}{
		"attorney": {
			shareCode: sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, LpaUID: "lpa-uid"},
			session:   &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
		},
		"replacement": {
			shareCode:     sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, LpaUID: "lpa-uid"},
			session:       &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement: true,
		},
		"trust corporation": {
			shareCode:          sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			shareCode:          sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"abcd1234"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			newCtx := mock.MatchedBy(func(ctx context.Context) bool {
				session, _ := appcontext.SessionFromContext(ctx)

				return assert.Equal(t, &appcontext.Session{SessionID: "aGV5", LpaID: "lpa-id"}, session)
			})

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
				Return(tc.shareCode, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Create(newCtx, tc.shareCode, "a@example.com").
				Return(&attorneydata.Provided{}, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Sub: "hey", Email: "a@example.com"}, nil)
			sessionStore.EXPECT().
				SetLogin(r, w, &sesh.LoginSession{
					Sub:     "hey",
					Email:   "a@example.com",
					HasLPAs: true,
				}).
				Return(nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(r.Context(), "lpa-uid").
				Return(nil, lpastore.ErrNotFound)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendMetric(newCtx, event.CategoryFunnelStartRate, event.MeasureOnlineAttorney).
				Return(nil)

			err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore, lpaStoreClient, eventClient)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathCodeOfConduct.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberWhenAttorneyAlreadySubmittedOnPaper(t *testing.T) {
	testcases := map[string]struct {
		shareCode          sharecodedata.Link
		session            *sesh.LoginSession
		isReplacement      bool
		isTrustCorporation bool
	}{
		"attorney": {
			shareCode: sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, LpaUID: "lpa-uid"},
			session:   &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
		},
		"replacement": {
			shareCode:     sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, LpaUID: "lpa-uid"},
			session:       &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement: true,
		},
		"trust corporation": {
			shareCode:          sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			shareCode:          sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, IsTrustCorporation: true, LpaUID: "lpa-uid"},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"abcd1234"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
				Return(tc.shareCode, nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(r.Context(), "lpa-uid").
				Return(&lpadata.Lpa{Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{UID: testUID, Channel: lpadata.ChannelPaper, SignedAt: &testNow}},
				}}, nil)
			lpaStoreClient.EXPECT().
				SendPaperAttorneyAccessOnline(r.Context(), "lpa-uid", "a@example.com", testUID).
				Return(nil)

			testAppData.LoginSessionEmail = "a@example.com"
			err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, lpaStoreClient, nil)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.PathDashboard.Format(), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberOnShareCodeStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {" abcd1234  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnShareCodeStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd 1-234 "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "abcd1234", ReferenceNumberRaw: "abcd 1-234"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, shareCodeStore, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnLpaStoreLpaError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, lpaStoreClient, nil)(testAppData, w, r)

	resp := w.Result()

	assert.ErrorContains(t, err, "error getting LPA from LPA store: err")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnLpaStoreSendPaperAttorneyAccessOnlineError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{ActorUID: testUID}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{Attorneys: lpadata.Attorneys{
			Attorneys: []lpadata.Attorney{{UID: testUID, Channel: lpadata.ChannelPaper, SignedAt: &testNow}}},
		}, nil)
	lpaStoreClient.EXPECT().
		SendPaperAttorneyAccessOnline(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, lpaStoreClient, nil)(testAppData, w, r)

	resp := w.Result()

	assert.ErrorContains(t, err, "error sending attorney email to LPA store: err")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnSessionGetError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(nil, lpastore.ErrNotFound)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil, lpaStoreClient, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterReferenceNumberOnSessionSetError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, lpastore.ErrNotFound)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil, lpaStoreClient, nil)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterReferenceNumberOnAttorneyStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, sharecodedata.HashedFromString("abcd1234")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(nil, lpastore.ErrNotFound)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&attorneydata.Provided{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore, lpaStoreClient, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnEventClientError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcd1234"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), LpaUID: "lpa-uid"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, lpastore.ErrNotFound)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&attorneydata.Provided{}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)
	sessionStore.EXPECT().
		SetLogin(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore, lpaStoreClient, eventClient)(testAppData, w, r)

	assert.ErrorIs(t, err, expectedError)
}

func TestPostEnterReferenceNumberOnValidationError(t *testing.T) {
	form := url.Values{
		"reference-number": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{},
		Errors: validation.With("reference-number", validation.EnterError{Label: "twelveCharactersReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateEnterReferenceNumberForm(t *testing.T) {
	testCases := map[string]struct {
		form   *enterReferenceNumberForm
		errors validation.List
	}{
		"valid": {
			form:   &enterReferenceNumberForm{ReferenceNumber: "abcd1234"},
			errors: nil,
		},
		"too short": {
			form: &enterReferenceNumberForm{ReferenceNumber: "1"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 8,
			}),
		},
		"too long": {
			form: &enterReferenceNumberForm{ReferenceNumber: "123456789ABCD"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 8,
			}),
		},
		"empty": {
			form: &enterReferenceNumberForm{},
			errors: validation.With("reference-number", validation.EnterError{
				Label: "twelveCharactersReferenceNumber",
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}

}
