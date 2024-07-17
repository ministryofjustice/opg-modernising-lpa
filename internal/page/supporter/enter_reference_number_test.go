package supporter

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

	err := EnterReferenceNumber(nil, template.Execute, nil, nil)(testAppData, w, r)
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

	err := EnterReferenceNumber(nil, template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	invite := &actor.MemberInvite{
		ReferenceNumber:  "abcd12345678",
		OrganisationID:   "org-id",
		OrganisationName: "org name",
		CreatedAt:        time.Now().Add(-47 * time.Hour),
		Email:            "a@example.org",
	}

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(r.Context()).
		Return(invite, nil)

	memberStore.EXPECT().
		CreateFromInvite(r.Context(), invite).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, &sesh.LoginSession{Email: "name@example.com", Sub: "a-sub", OrganisationID: "org-id", OrganisationName: "org name"}).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "member invite redeemed", slog.String("organisationID", "org-id"))

	err := EnterReferenceNumber(logger, nil, memberStore, sessionStore)(testAppData, w, r)
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
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

	err := EnterReferenceNumber(nil, template.Execute, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenInviteExpired(t *testing.T) {
	form := url.Values{"reference-number": {"match-1234-789"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(r.Context()).
		Return(&actor.MemberInvite{
			ReferenceNumber: "match1234789",
			OrganisationID:  "org-id",
			CreatedAt:       time.Now().Add(-49 * time.Hour),
		}, nil)

	err := EnterReferenceNumber(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Supporter.InviteExpired.Format(), resp.Header.Get("Location"))
}

func TestPostEnterReferenceNumberWhenMemberStoreInvitedMemberError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{}, expectedError)

	err := EnterReferenceNumber(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenMemberStoreCreateError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{
			ReferenceNumber: "abcd12345678",
			OrganisationID:  "org-id",
			CreatedAt:       time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterReferenceNumber(nil, nil, memberStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberWhenSessionGetError(t *testing.T) {
	form := url.Values{"reference-number": {"abcd12345678"}}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{
			ReferenceNumber: "abcd12345678",
			OrganisationID:  "org-id",
			CreatedAt:       time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(nil, expectedError)

	err := EnterReferenceNumber(nil, nil, memberStore, sessionStore)(testAppData, w, r)
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

	memberStore := newMockMemberStore(t)
	memberStore.EXPECT().
		InvitedMember(mock.Anything).
		Return(&actor.MemberInvite{
			ReferenceNumber: "abcd12345678",
			OrganisationID:  "org-id",
			CreatedAt:       time.Now(),
		}, nil)

	memberStore.EXPECT().
		CreateFromInvite(mock.Anything, mock.Anything).
		Return(nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{Email: "name@example.com", Sub: "a-sub"}, nil)
	sessionStore.EXPECT().
		SetLogin(r, w, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	err := EnterReferenceNumber(logger, nil, memberStore, sessionStore)(testAppData, w, r)
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

	err := EnterReferenceNumber(nil, template.Execute, nil, nil)(testAppData, w, r)
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
