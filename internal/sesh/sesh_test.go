package sesh

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sessions "github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
)

var (
	expectedError = errors.New("err")
)

func (m *mockCookieStore) expectGet(r *http.Request, cookie, param string, values any) *mockCookieStore {
	session := sessions.NewSession(m, cookie)
	session.Values = map[any]any{param: values}

	m.EXPECT().
		Get(r, cookie).
		Return(session, nil)

	return m
}

func (m *mockCookieStore) expectSet(r *http.Request, w http.ResponseWriter, cookie, param string, values any, options *sessions.Options) *mockCookieStore {
	session := sessions.NewSession(m, cookie)
	session.Options = options
	session.Values = map[any]any{param: values}

	m.EXPECT().
		Save(r, w, session).
		Return(expectedError)

	return m
}

func (m *mockCookieStore) expectClear(r *http.Request, w http.ResponseWriter, cookie string) *mockCookieStore {
	session := sessions.NewSession(m, cookie)
	session.Options.MaxAge = -1
	session.Values = map[any]any{}

	m.EXPECT().
		Get(r, cookie).
		Return(session, nil)
	m.EXPECT().
		Save(r, w, session).
		Return(expectedError)

	return m
}

func TestOneLogin(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		values = &OneLoginSession{State: "a", Nonce: "b", Redirect: "c", SessionID: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectGet(r, cookieSignIn, "one-login", values)

	store := &Store{s: cookieStore}

	result, err := store.OneLogin(r)
	assert.Nil(t, err)
	assert.Equal(t, values, result)
}

func TestSetOneLogin(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w      = httptest.NewRecorder()
		values = &OneLoginSession{SessionID: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectSet(r, w, cookieSignIn, "one-login", values, oneLoginCookieOptions)

	store := &Store{s: cookieStore}

	err := store.SetOneLogin(r, w, &OneLoginSession{SessionID: "x"})
	assert.Equal(t, expectedError, err)
}

func TestLogin(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		values = &LoginSession{Sub: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectGet(r, cookieSession, "session", values)

	store := &Store{d: cookieStore}

	result, err := store.Login(r)
	assert.Nil(t, err)
	assert.Equal(t, values, result)
}

func TestSetLogin(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w      = httptest.NewRecorder()
		values = &LoginSession{Sub: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectSet(r, w, cookieSession, "session", values, sessionCookieOptions)

	store := &Store{d: cookieStore}

	err := store.SetLogin(r, w, values)
	assert.Equal(t, expectedError, err)
}

func TestClearLogin(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()
	)

	cookieStore := newMockCookieStore(t).
		expectClear(r, w, cookieSession)

	store := &Store{d: cookieStore}

	err := store.ClearLogin(r, w)
	assert.Equal(t, expectedError, err)
}

func TestPayment(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		values = &PaymentSession{PaymentID: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectGet(r, cookiePayment, "payment", values)

	store := &Store{s: cookieStore}

	result, err := store.Payment(r)
	assert.Nil(t, err)
	assert.Equal(t, values, result)
}

func TestSetPayment(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w      = httptest.NewRecorder()
		values = &PaymentSession{PaymentID: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectSet(r, w, cookiePayment, "payment", values, paymentCookieOptions)

	store := &Store{s: cookieStore}

	err := store.SetPayment(r, w, values)
	assert.Equal(t, expectedError, err)
}

func TestClearPayment(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()
	)

	cookieStore := newMockCookieStore(t).
		expectClear(r, w, cookiePayment)

	store := &Store{s: cookieStore}

	err := store.ClearPayment(r, w)
	assert.Equal(t, expectedError, err)
}

func TestCsrf(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		values = &CsrfSession{Token: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectGet(r, cookieCsrf, "csrf", values)

	store := &Store{s: cookieStore}

	result, err := store.Csrf(r)
	assert.Nil(t, err)
	assert.Equal(t, values, result)
}

func TestSetCsrf(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w      = httptest.NewRecorder()
		values = &CsrfSession{Token: "x"}
	)

	cookieStore := newMockCookieStore(t).
		expectSet(r, w, cookieCsrf, "csrf", values, sessionCookieOptions)

	store := &Store{s: cookieStore}

	err := store.SetCsrf(r, w, values)
	assert.Equal(t, expectedError, err)
}

func TestLpaData(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		values = &LpaDataSession{LpaID: "lpa-id"}
	)

	cookieStore := newMockCookieStore(t).
		expectGet(r, cookieLPAData, "lpa-data", values)

	store := &Store{s: cookieStore}

	result, err := store.LpaData(r)
	assert.Nil(t, err)
	assert.Equal(t, values, result)
}

func TestSetLpaDataSession(t *testing.T) {
	var (
		r, _   = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w      = httptest.NewRecorder()
		values = &LpaDataSession{LpaID: "lpa-id"}
	)

	cookieStore := newMockCookieStore(t).
		expectSet(r, w, cookieLPAData, "lpa-data", values, sessionCookieOptions)

	store := &Store{s: cookieStore}

	err := store.SetLpaData(r, w, values)
	assert.Equal(t, expectedError, err)
}

func TestLpaDataSessionValid(t *testing.T) {
	valid := &LpaDataSession{LpaID: "lpa-id"}
	assert.True(t, valid.Valid())

	invalid := &LpaDataSession{}
	assert.False(t, invalid.Valid())
}
