package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestWhoIsEligible(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity: true,
			LpaID:    "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(&page.Lpa{ID: "lpa-id"}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whoIsEligibleData{
			Lpa: &page.Lpa{ID: "lpa-id"},
			App: testAppData,
		}).
		Return(nil)

	err := WhoIsEligible(template.Execute, lpaStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWhoIsEligibleWhenSessionStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "shareCode").
		Return(&sessions.Session{}, expectedError)

	err := WhoIsEligible(nil, nil, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWhoIsEligibleWhenLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)

	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity: true,
			LpaID:    "lpa-id",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(&page.Lpa{}, expectedError)

	err := WhoIsEligible(nil, lpaStore, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
