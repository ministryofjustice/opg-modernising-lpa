package attorney

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockSessionStore) ExpectGet(r *http.Request, values map[any]any, err error) {
	session := sessions.NewSession(m, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = values
	m.EXPECT().
		Get(r, "session").
		Return(session, err)
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
			shareCode: actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id"},
			session:   &sesh.LoginSession{Sub: "hey"},
		},
		"replacement": {
			shareCode:     actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id", IsReplacementAttorney: true},
			session:       &sesh.LoginSession{Sub: "hey"},
			isReplacement: true,
		},
		"trust corporation": {
			shareCode:          actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id", IsTrustCorporation: true},
			session:            &sesh.LoginSession{Sub: "hey"},
			isTrustCorporation: true,
		},
		"replacement trust corporation": {
			shareCode:          actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id", IsReplacementAttorney: true, IsTrustCorporation: true},
			session:            &sesh.LoginSession{Sub: "hey"},
			isReplacement:      true,
			isTrustCorporation: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"a Ref-Number12"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.EXPECT().
				Get(r.Context(), actor.TypeAttorney, "aRefNumber12").
				Return(tc.shareCode, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Create(mock.MatchedBy(func(ctx context.Context) bool {
					session, _ := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "aGV5", LpaID: "lpa-id"}, session)
				}), "aGV5", "attorney-id", tc.isReplacement, tc.isTrustCorporation).
				Return(&actor.AttorneyProvidedDetails{}, nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.
				ExpectGet(r,
					map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, nil)

			err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, attorneyStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.Attorney.CodeOfConduct.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberOnDonorStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"  aRefNumber12  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnShareCodeStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"a Ref-Number12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "aRefNumber12", ReferenceNumberRaw: "a Ref-Number12"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, shareCodeStore, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnSessionGetError(t *testing.T) {
	form := url.Values{
		"reference-number": {"aRefNumber12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		ExpectGet(r,
			map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostEnterReferenceNumberOnAttorneyStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"a Ref-Number12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Create(mock.Anything, mock.Anything, mock.Anything, false, false).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		ExpectGet(r,
			map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, nil)

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
