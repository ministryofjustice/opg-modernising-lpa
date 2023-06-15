package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestWhoIsEligible(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity:        true,
			LpaID:           "lpa-id",
			DonorFullName:   "Full name",
			DonorFirstNames: "Full",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whoIsEligibleData{
			DonorFullName:   "Full name",
			DonorFirstNames: "Full",
			App:             testAppData,
		}).
		Return(nil)

	err := WhoIsEligible(template.Execute, sessionStore)(testAppData, w, r)
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

	err := WhoIsEligible(nil, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWhoIsEligibleOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	shareCodeSession := sessions.NewSession(sessionStore, "shareCode")
	shareCodeSession.Values = map[any]any{
		"share-code": &sesh.ShareCodeSession{
			Identity:        true,
			LpaID:           "lpa-id",
			DonorFullName:   "Full name",
			DonorFirstNames: "Full",
		},
	}

	sessionStore.
		On("Get", r, "shareCode").
		Return(shareCodeSession, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, whoIsEligibleData{
			DonorFullName:   "Full name",
			DonorFirstNames: "Full",
			App:             testAppData,
		}).
		Return(expectedError)

	err := WhoIsEligible(template.Execute, sessionStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
