package attorney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
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

	err := EnterReferenceNumber(template.Execute, nil, nil, nil)(testAppData, w, r)

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

	err := EnterReferenceNumber(template.Execute, nil, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	testcases := map[string]struct {
		shareCode          actor.ShareCodeData
		session            *sesh.LoginSession
		isReplacement      bool
		isTrustCorporation bool
	}{
		"attorney": {
			shareCode: actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID},
			session:   &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
		},
		"replacement": {
			shareCode:     actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true},
			session:       &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement: true,
		},
		"trust corporation": {
			shareCode:          actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsTrustCorporation: true},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			shareCode:          actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID, IsReplacementAttorney: true, IsTrustCorporation: true},
			session:            &sesh.LoginSession{Sub: "hey", Email: "a@example.com"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"abcdef123456"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, "abcdef123456").
				Return(tc.shareCode, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Create(mock.MatchedBy(func(ctx context.Context) bool {
					session, _ := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "aGV5", LpaID: "lpa-id"}, session)
				}), tc.shareCode, "a@example.com").
				Return(&actor.AttorneyProvidedDetails{}, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				Login(r).
				Return(&sesh.LoginSession{Sub: "hey", Email: "a@example.com"}, nil)

			err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.CodeOfConduct.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberWhenConditionalCheckFailed(t *testing.T) {
	shareCode := actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("")), ActorUID: testUID}

	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), mock.Anything, mock.Anything).
		Return(shareCode, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&actor.AttorneyProvidedDetails{}, dynamo.ConditionalCheckFailedError{})

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey", Email: "a@example.com"}, nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.CodeOfConduct.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostEnterReferenceNumberOnDonorStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {" abcdef123456  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "abcdef123456").
		Return(actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnShareCodeStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcde f-123456 "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "abcdef123456", ReferenceNumberRaw: "abcde f-123456"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "abcdef123456").
		Return(actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, shareCodeStore, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnSessionGetError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "abcdef123456").
		Return(actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostEnterReferenceNumberOnAttorneyStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "abcdef123456").
		Return(actor.ShareCodeData{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey(""))}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Sub: "hey"}, nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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

	err := EnterReferenceNumber(template.Execute, nil, nil, nil)(testAppData, w, r)

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
			form: &enterReferenceNumberForm{ReferenceNumber: "1"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 12,
			}),
		},
		"too long": {
			form: &enterReferenceNumberForm{ReferenceNumber: "123456789ABCD"},
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
