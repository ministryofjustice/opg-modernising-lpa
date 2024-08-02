// Package sesh provides functionality for setting and reading session data as cookies.
package sesh

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// These are the cookie names in use. We need some to be able to overlap
// (e.g. session+pay, so you can be signed in and pay for something), but others
// shouldn't (i.e. the reuse of session as you can't be signed in twice).
const (
	cookieCsrf    = "csrf"
	cookieLPAData = "lpa-data"
	cookiePayment = "pay"
	cookieSession = "session"
	cookieSignIn  = "params"
)

var (
	sessionCookieOptions = &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	oneLoginCookieOptions = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	paymentCookieOptions = &sessions.Options{
		Path: "/",
		// A payment can be resumed up to 90 minutes after creation
		MaxAge:   int(time.Minute * 90 / time.Second),
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
)

func init() {
	gob.Register(&OneLoginSession{})
	gob.Register(&LoginSession{})
	gob.Register(&PaymentSession{})
	gob.Register(&CsrfSession{})
	gob.Register(&LpaDataSession{})
}

type cookieStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

type MissingSessionError string

func (e MissingSessionError) Error() string {
	return fmt.Sprintf("missing %s session", string(e))
}

type InvalidSessionError string

func (e InvalidSessionError) Error() string {
	return fmt.Sprintf("%s session invalid", string(e))
}

type Store struct {
	s sessions.Store
}

func NewStore(keyPairs [][]byte) *Store {
	return &Store{s: sessions.NewCookieStore(keyPairs...)}
}

type isValid interface {
	Valid() bool
}

func getSession[T isValid](s sessions.Store, cookieName, paramKey string, r *http.Request) (T, error) {
	params, err := s.Get(r, cookieName)
	if err != nil {
		return *new(T), err
	}

	value, ok := params.Values[paramKey]
	if !ok {
		return *new(T), MissingSessionError(paramKey)
	}

	session, ok := value.(T)
	if !ok {
		return *new(T), MissingSessionError(paramKey)
	}
	if !session.Valid() {
		return *new(T), InvalidSessionError(paramKey)
	}

	return session, nil
}

func setSession[T any](s sessions.Store, cookieName, paramKey string, r *http.Request, w http.ResponseWriter, values T, options *sessions.Options) error {
	session := sessions.NewSession(s, cookieName)
	session.Values = map[any]any{paramKey: values}
	session.Options = options
	return s.Save(r, w, session)
}

func clearSession(s sessions.Store, cookieName string, r *http.Request, w http.ResponseWriter) error {
	session, err := s.Get(r, cookieName)
	if err != nil {
		return err
	}
	session.Values = map[any]any{}
	session.Options.MaxAge = -1
	return s.Save(r, w, session)
}

type OneLoginSession struct {
	State     string
	Nonce     string
	Locale    string
	Redirect  string
	SessionID string
	LpaID     string
}

func (s OneLoginSession) Valid() bool {
	return s.State != "" && s.Nonce != "" && s.Redirect != ""
}

func (s *Store) OneLogin(r *http.Request) (*OneLoginSession, error) {
	return getSession[*OneLoginSession](s.s, cookieSignIn, "one-login", r)
}

func (s *Store) SetOneLogin(r *http.Request, w http.ResponseWriter, oneLoginSession *OneLoginSession) error {
	return setSession(s.s, cookieSignIn, "one-login", r, w, oneLoginSession, oneLoginCookieOptions)
}

type LoginSession struct {
	IDToken          string
	Sub              string
	Email            string
	OrganisationID   string
	OrganisationName string
}

func (s LoginSession) SessionID() string {
	return base64.StdEncoding.EncodeToString([]byte(s.Sub))
}

func (s LoginSession) Valid() bool {
	return s.Sub != ""
}

func (s *Store) Login(r *http.Request) (*LoginSession, error) {
	return getSession[*LoginSession](s.s, cookieSession, "session", r)
}

func (s *Store) SetLogin(r *http.Request, w http.ResponseWriter, donorSession *LoginSession) error {
	return setSession(s.s, cookieSession, "session", r, w, donorSession, sessionCookieOptions)
}

func (s *Store) ClearLogin(r *http.Request, w http.ResponseWriter) error {
	return clearSession(s.s, cookieSession, r, w)
}

type PaymentSession struct {
	PaymentID string
}

func (s PaymentSession) Valid() bool {
	return true
}

func (s *Store) Payment(r *http.Request) (*PaymentSession, error) {
	return getSession[*PaymentSession](s.s, cookiePayment, "payment", r)
}

func (s *Store) SetPayment(r *http.Request, w http.ResponseWriter, paymentSession *PaymentSession) error {
	return setSession(s.s, cookiePayment, "payment", r, w, paymentSession, paymentCookieOptions)
}

func (s *Store) ClearPayment(r *http.Request, w http.ResponseWriter) error {
	return clearSession(s.s, cookiePayment, r, w)
}

type CsrfSession struct {
	IsNew bool
	Token string
}

func (s CsrfSession) Valid() bool {
	return true
}

func (s *Store) Csrf(r *http.Request) (*CsrfSession, error) {
	return getSession[*CsrfSession](s.s, cookieCsrf, "csrf", r)
}

func (s *Store) SetCsrf(r *http.Request, w http.ResponseWriter, csrfSession *CsrfSession) error {
	return setSession(s.s, cookieCsrf, "csrf", r, w, csrfSession, sessionCookieOptions)
}

type LpaDataSession struct {
	LpaID string
}

func (s LpaDataSession) Valid() bool {
	return s.LpaID != ""
}

func (s *Store) LpaData(r *http.Request) (*LpaDataSession, error) {
	return getSession[*LpaDataSession](s.s, cookieLPAData, "lpa-data", r)
}

func (s *Store) SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *LpaDataSession) error {
	return setSession(s.s, cookieLPAData, "lpa-data", r, w, lpaDataSession, sessionCookieOptions)
}
