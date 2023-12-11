package page

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSignOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Options: &sessions.Options{},
			Values:  map[any]any{"session": &sesh.LoginSession{IDToken: "id-token", Sub: "abc"}},
		}, nil)
	sessionStore.
		On("Save", r, w, &sessions.Session{
			Options: &sessions.Options{MaxAge: -1},
			Values:  map[any]any{},
		}).
		Return(nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("EndSessionURL", "id-token", "http://public"+Paths.Start.Format()).
		Return("http://end-session", nil)

	err := SignOut(nil, sessionStore, oneLoginClient, "http://public")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://end-session", resp.Header.Get("Location"))
}

func TestSignOutWhenEndSessionURLFails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", "unable to end onelogin session: err")

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Options: &sessions.Options{},
			Values:  map[any]any{"session": &sesh.LoginSession{IDToken: "id-token", Sub: "abc"}},
		}, nil)
	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("EndSessionURL", "id-token", "http://public"+Paths.Start.Format()).
		Return("", expectedError)

	err := SignOut(logger, sessionStore, oneLoginClient, "http://public")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://public"+Paths.Start.Format(), resp.Header.Get("Location"))
}

func TestSignOutWhenClearSessionFails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	logger := newMockLogger(t)
	logger.
		On("Print", "unable to expire session: err")

	sessionStore := newMockSessionStore(t)
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{
			Options: &sessions.Options{},
			Values:  map[any]any{"session": &sesh.LoginSession{IDToken: "id-token", Sub: "abc"}},
		}, nil)
	sessionStore.
		On("Save", r, w, mock.Anything).
		Return(errors.New("err"))

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.
		On("EndSessionURL", "id-token", "http://public"+Paths.Start.Format()).
		Return("http://end-session", nil)

	err := SignOut(logger, sessionStore, oneLoginClient, "http://public")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://end-session", resp.Header.Get("Location"))
}
