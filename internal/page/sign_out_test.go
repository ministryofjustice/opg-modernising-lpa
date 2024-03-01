package page

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
)

func TestSignOut(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{IDToken: "id-token", Sub: "abc"}, nil)
	sessionStore.EXPECT().
		ClearLogin(r, w).
		Return(nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		EndSessionURL("id-token", "http://public"+Paths.Start.Format()).
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
	logger.EXPECT().
		Info("unable to end onelogin session", slog.Any("err", expectedError))

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{IDToken: "id-token", Sub: "abc"}, nil)
	sessionStore.EXPECT().
		ClearLogin(r, w).
		Return(nil)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		EndSessionURL("id-token", "http://public"+Paths.Start.Format()).
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
	logger.EXPECT().
		Info("unable to expire session", slog.Any("err", expectedError))

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		Login(r).
		Return(&sesh.LoginSession{IDToken: "id-token", Sub: "abc"}, nil)
	sessionStore.EXPECT().
		ClearLogin(r, w).
		Return(expectedError)

	oneLoginClient := newMockOneLoginClient(t)
	oneLoginClient.EXPECT().
		EndSessionURL("id-token", "http://public"+Paths.Start.Format()).
		Return("http://end-session", nil)

	err := SignOut(logger, sessionStore, oneLoginClient, "http://public")(TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://end-session", resp.Header.Get("Location"))
}
