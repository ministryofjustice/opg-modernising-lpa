package supporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterReferenceNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterReferenceNumber{
			App: testAppData,
			Form: &referenceNumberForm{
				Label: "referenceNumber",
			},
		}).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReferenceNumberWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterReferenceNumber{
			App: testAppData,
			Form: &referenceNumberForm{
				Label: "referenceNumber",
			},
		}).
		Return(expectedError)

	err := EnterReferenceNumber(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(r.Context()).
		Return(&actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id", OrganisationName: "org name"}, nil)

	organisationStore.EXPECT().
		CreateMember(r.Context(), &actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id", OrganisationName: "org name"}).
		Return(nil)

	sessionStore := newMockSessionStore(t)

	session := sessions.NewSession(sessionStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"session": &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}}

	sessionStore.EXPECT().
		Get(r, "session").
		Return(session, nil)

	session.Values = map[any]any{"session": &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub", OrganisationID: "org-id", OrganisationName: "org name"}}

	sessionStore.EXPECT().
		Save(r, w, session).
		Return(nil)

	err := EnterReferenceNumber(nil, organisationStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Dashboard.Format(), resp.Header.Get("Location"))
}

func TestPostEnterReferenceNumberWhenIncorrectReferenceNumber(t *testing.T) {
	form := url.Values{"reference-number": {"not-match-1234"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(r.Context()).
		Return(&actor.MemberInvite{ReferenceNumber: "notmatch123", OrganisationID: "org-id"}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterReferenceNumber{
			App: testAppData,
			Form: &referenceNumberForm{
				Label:              "referenceNumber",
				ReferenceNumber:    "notmatch1234",
				ReferenceNumberRaw: "not-match-1234",
			},
			Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
		}).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, organisationStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenOrganisationStoreInvitedMemberError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id"}, expectedError)

	err := EnterReferenceNumber(nil, organisationStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenOrganisationStoreCreateError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id"}, nil)

	organisationStore.EXPECT().
		CreateMember(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, organisationStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenSessionGetError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id"}, nil)

	organisationStore.EXPECT().
		CreateMember(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)

	session := sessions.NewSession(sessionStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"session": &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}}

	sessionStore.EXPECT().
		Get(r, mock.Anything).
		Return(session, expectedError)

	err := EnterReferenceNumber(nil, organisationStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.Start.Format(), resp.Header.Get("Location"))
}

func TestPostEnterReferenceNumberWhenSessionSaveError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	organisationStore := newMockOrganisationStore(t)
	organisationStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{ReferenceNumber: "abcd12345678", OrganisationID: "org-id"}, nil)

	organisationStore.EXPECT().
		CreateMember(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)

	session := sessions.NewSession(sessionStore, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	session.Values = map[any]any{"session": &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}}

	sessionStore.EXPECT().
		Get(r, mock.Anything).
		Return(session, nil)

	session.Values = map[any]any{"session": &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub", OrganisationID: "org-id"}}

	sessionStore.EXPECT().
		Save(r, w, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, organisationStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenValidationError(t *testing.T) {
	form := url.Values{"reference-number": {""}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &enterReferenceNumber{
			App: testAppData,
			Form: &referenceNumberForm{
				Label: "referenceNumber",
			},
			Errors: validation.With("reference-number", validation.EnterError{Label: "twelveCharactersReferenceNumber"}),
		}).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReferenceNumberFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *referenceNumberForm
		errors validation.List
	}{
		"valid": {
			form:   &referenceNumberForm{ReferenceNumber: "abcdef123456"},
			errors: nil,
		},
		"too short": {
			form: &referenceNumberForm{ReferenceNumber: "abcdef12345"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 12,
			}),
		},
		"too long": {
			form: &referenceNumberForm{ReferenceNumber: "abcdef1234567"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "theReferenceNumberYouEnter",
				Length: 12,
			}),
		},
		"empty": {
			form: &referenceNumberForm{},
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
