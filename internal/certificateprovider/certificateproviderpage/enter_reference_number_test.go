package certificateproviderpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	err := EnterReferenceNumber(template.Execute, newMockShareCodeStore(t), nil, nil, nil)(testAppData, w, r)

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

	err := EnterReferenceNumber(template.Execute, newMockShareCodeStore(t), nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef 123-456"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeData := sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}
	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeCertificateProvider, sharecodedata.HashedFromString("abcdef123456")).
		Return(shareCodeData, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(r.Context(), "lpa-uid").
		Return(nil, lpastore.ErrNotFound)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey", Email: "a@b.com"}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Create(mock.MatchedBy(func(ctx context.Context) bool {
			session, _ := appcontext.SessionFromContext(ctx)

			return assert.Equal(t, &appcontext.Session{SessionID: "aGV5", LpaID: "lpa-id"}, session)
		}), shareCodeData, "a@b.com").
		Return(&certificateproviderdata.Provided{}, nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, certificateProviderStore, lpaStoreClient)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathWhoIsEligible.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterReferenceNumberWhenPaperCertificateExists(t *testing.T) {
	testcases := map[string]struct {
		redirectURL string
		sessionID   string
	}{
		"logged in": {
			redirectURL: page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn.Format() + "?donorFullName=a+b&lpaType=c",
			sessionID:   "a",
		},
		"not logged in": {
			redirectURL: page.PathCertificateProviderYouHaveAlreadyProvidedACertificate.Format() + "?donorFullName=a+b&lpaType=c",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"abcdef 123-456"},
			}

			uid := actoruid.New()
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(mock.Anything, mock.Anything, mock.Anything).
				Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}, nil)

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(r.Context(), "lpa-uid").
				Return(&lpadata.Lpa{
					Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
					CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper, SignedAt: &testNow},
					Type:                lpadata.LpaTypePropertyAndAffairs,
				}, nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("c")
			testAppData.Localizer = localizer
			testAppData.SessionID = tc.sessionID

			err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, lpaStoreClient)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.redirectURL, resp.Header.Get("Location"))
		})
	}

}

func TestPostEnterReferenceNumberOnShareCodeStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeCertificateProvider, sharecodedata.HashedFromString("abcdef123456")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id"))}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.ErrorContains(t, err, "error getting shareCode: err")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnShareCodeStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef 123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "abcdef123456", ReferenceNumberRaw: "abcdef 123456"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeCertificateProvider, sharecodedata.HashedFromString("abcdef123456")).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id"))}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, shareCodeStore, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenLpaStoreClientError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef 123-456"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, lpaStoreClient)(testAppData, w, r)
	assert.ErrorContains(t, err, "error getting LPA from LPA store: err")
}

func TestPostEnterReferenceNumberWhenCreateError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef 123-456"},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(mock.Anything, mock.Anything, mock.Anything).
		Return(sharecodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("session-id")), ActorUID: uid, LpaUID: "lpa-uid"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(mock.Anything).
		Return(&sesh.LoginSession{Sub: "hey", Email: "a@b.com"}, nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(nil, lpastore.ErrNotFound)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	testAppData.SessionID = ""

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, certificateProviderStore, lpaStoreClient)(testAppData, w, r)
	assert.ErrorContains(t, err, "error creating certificate provider: err")
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

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil)(testAppData, w, r)

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
			form:   &enterReferenceNumberForm{ReferenceNumber: "abcdef123456"},
			errors: nil,
		},
		"too short": {
			form: &enterReferenceNumberForm{ReferenceNumber: "abcdef12345"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 12,
			}),
		},
		"too long": {
			form: &enterReferenceNumberForm{ReferenceNumber: "abcdef1234567"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 12,
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
