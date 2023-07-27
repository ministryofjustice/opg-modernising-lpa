package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
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
	m.
		On("Get", r, "session").
		Return(session, err)
}

func TestGetEnterReferenceNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  TestAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReferenceNumberOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  TestAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(expectedError)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	testcases := map[actor.Type]struct {
		shareCode        actor.ShareCodeData
		session          *sesh.LoginSession
		isReplacement    bool
		expectedRedirect string
	}{
		actor.TypeAttorney: {
			shareCode:        actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id"},
			session:          &sesh.LoginSession{Sub: "hey"},
			expectedRedirect: Paths.Attorney.CodeOfConduct.Format("lpa-id"),
		},
		actor.TypeReplacementAttorney: {
			shareCode:        actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", AttorneyID: "attorney-id", IsReplacementAttorney: true},
			session:          &sesh.LoginSession{Sub: "hey"},
			isReplacement:    true,
			expectedRedirect: Paths.Attorney.CodeOfConduct.Format("lpa-id"),
		},
		actor.TypeCertificateProvider: {
			shareCode:        actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"},
			session:          &sesh.LoginSession{Sub: "hey"},
			expectedRedirect: Paths.CertificateProvider.WhoIsEligible.Format("lpa-id"),
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			form := url.Values{
				"reference-number": {"a Ref-Number12"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", FormUrlEncoded)

			shareCodeStore := newMockShareCodeStore(t)
			shareCodeStore.
				On("Get", r.Context(), actorType, "aRefNumber12").
				Return(tc.shareCode, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("Create", mock.MatchedBy(func(ctx context.Context) bool {
					session, _ := SessionDataFromContext(ctx)

					return assert.Equal(t, &SessionData{SessionID: "aGV5", LpaID: "lpa-id"}, session)
				}), "aGV5", "attorney-id", tc.isReplacement).
				Return(&actor.AttorneyProvidedDetails{}, nil).
				Maybe()

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.
				On("Create", mock.MatchedBy(func(ctx context.Context) bool {
					session, _ := SessionDataFromContext(ctx)

					return assert.Equal(t, &SessionData{SessionID: "aGV5", LpaID: "lpa-id"}, session)
				}), "aGV5").
				Return(&actor.CertificateProviderProvidedDetails{}, nil).
				Maybe()

			sessionStore := newMockSessionStore(t)
			sessionStore.
				ExpectGet(r,
					map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, nil)

			err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, certificateProviderStore, attorneyStore, actorType)(TestAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberOnDonorStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"  aRefNumber12  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Get", r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", Identity: true}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, nil, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

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
	r.Header.Add("Content-Type", FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    TestAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "aRefNumber12", ReferenceNumberRaw: "a Ref-Number12"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Get", r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", Identity: true}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, shareCodeStore, nil, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

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
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Get", r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", Identity: true}, nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		ExpectGet(r,
			map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, expectedError)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestPostEnterReferenceNumberOnAttorneyStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"a Ref-Number12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Get", r.Context(), actor.TypeAttorney, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", Identity: true}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Create", mock.Anything, mock.Anything, mock.Anything, false).
		Return(&actor.AttorneyProvidedDetails{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		ExpectGet(r,
			map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, nil, attorneyStore, actor.TypeAttorney)(TestAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnCertificateProviderStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"a Ref-Number12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.
		On("Get", r.Context(), actor.TypeCertificateProvider, "aRefNumber12").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", Identity: true}, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Create", mock.Anything, mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		ExpectGet(r,
			map[any]any{"session": &sesh.LoginSession{Sub: "hey"}}, nil)

	err := EnterReferenceNumber(nil, shareCodeStore, sessionStore, certificateProviderStore, nil, actor.TypeCertificateProvider)(TestAppData, w, r)

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
	r.Header.Add("Content-Type", FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    TestAppData,
		Form:   &enterReferenceNumberForm{},
		Errors: validation.With("reference-number", validation.EnterError{Label: "twelveCharactersReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil, nil, nil, actor.TypeAttorney)(TestAppData, w, r)

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
				Label:  "referenceNumberMustBeTwelveCharacters",
				Length: 12,
			}),
		},
		"too long": {
			form: &enterReferenceNumberForm{ReferenceNumber: "abcdef1234567"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "referenceNumberMustBeTwelveCharacters",
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
