package sesh

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

type Store = sessions.Store

type MissingSessionError string

func (e MissingSessionError) Error() string {
	return fmt.Sprintf("missing %s session", string(e))
}

type InvalidSessionError string

func (e InvalidSessionError) Error() string {
	return fmt.Sprintf("%s session invalid", string(e))
}

// These are the cookie names in use. We need some to be able to overlap
// (e.g. session+pay, so you can be signed in and pay for something), but others
// shouldn't (i.e. the reuse of session as you can't be signed in twice).
const (
	cookieSignIn    = "params"
	cookieSession   = "session"
	cookieYoti      = "yoti"
	cookiePay       = "pay"
	cookieShareCode = "shareCode"
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
	gob.Register(&YotiSession{})
	gob.Register(&LoginSession{})
	gob.Register(&PaymentSession{})
	gob.Register(&ShareCodeSession{})
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

func OneLogin(store sessions.Store, r *http.Request) (*OneLoginSession, error) {
	params, err := store.Get(r, cookieSignIn)
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["one-login"]
	if !ok {
		return nil, MissingSessionError("one-login")
	}

	oneLoginSession, ok := session.(*OneLoginSession)
	if !ok {
		return nil, MissingSessionError("one-login")
	}
	if !oneLoginSession.Valid() {
		return nil, InvalidSessionError("one-login")
	}

	return oneLoginSession, nil
}

func SetOneLogin(store sessions.Store, r *http.Request, w http.ResponseWriter, oneLoginSession *OneLoginSession) error {
	params := sessions.NewSession(store, cookieSignIn)
	params.Values = map[any]any{"one-login": oneLoginSession}
	params.Options = oneLoginCookieOptions
	return store.Save(r, w, params)
}

type YotiSession struct {
	Locale              string
	LpaID               string
	CertificateProvider bool
}

func (s YotiSession) Valid() bool {
	return s.LpaID != ""
}

func Yoti(store sessions.Store, r *http.Request) (*YotiSession, error) {
	params, err := store.Get(r, cookieYoti)
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["yoti"]
	if !ok {
		return nil, MissingSessionError("yoti")
	}

	yotiSession, ok := session.(*YotiSession)
	if !ok {
		return nil, MissingSessionError("yoti")
	}
	if !yotiSession.Valid() {
		return nil, InvalidSessionError("yoti")
	}

	return yotiSession, nil
}

func SetYoti(store sessions.Store, r *http.Request, w http.ResponseWriter, yotiSession *YotiSession) error {
	params := sessions.NewSession(store, cookieYoti)
	params.Values = map[any]any{"yoti": yotiSession}
	params.Options = oneLoginCookieOptions
	return store.Save(r, w, params)
}

type LoginSession struct {
	IDToken string
	Sub     string
	Email   string
}

func (s LoginSession) Valid() bool {
	return s.Sub != ""
}

func Login(store sessions.Store, r *http.Request) (*LoginSession, error) {
	params, err := store.Get(r, cookieSession)
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["session"]
	if !ok {
		return nil, MissingSessionError("session")
	}

	loginSession, ok := session.(*LoginSession)
	if !ok {
		return nil, MissingSessionError("session")
	}
	if !loginSession.Valid() {
		return nil, InvalidSessionError("session")
	}

	return loginSession, nil
}

func SetLoginSession(store sessions.Store, r *http.Request, w http.ResponseWriter, donorSession *LoginSession) error {
	session := sessions.NewSession(store, cookieSession)
	session.Values = map[any]any{"session": donorSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}

func ClearLoginSession(store Store, r *http.Request, w http.ResponseWriter) error {
	session, err := store.Get(r, cookieSession)
	if err != nil {
		return err
	}
	session.Values = map[any]any{}
	session.Options.MaxAge = -1
	return store.Save(r, w, session)
}

type PaymentSession struct {
	PaymentID string
}

func (s PaymentSession) Valid() bool {
	return true
}

func Payment(store sessions.Store, r *http.Request) (*PaymentSession, error) {
	params, err := store.Get(r, cookiePay)
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["payment"]
	if !ok {
		return nil, MissingSessionError("payment")
	}

	paymentSession, ok := session.(*PaymentSession)
	if !ok {
		return nil, MissingSessionError("payment")
	}
	if !paymentSession.Valid() {
		return nil, InvalidSessionError("payment")
	}

	return paymentSession, nil
}

func SetPayment(store sessions.Store, r *http.Request, w http.ResponseWriter, paymentSession *PaymentSession) error {
	session := sessions.NewSession(store, cookiePay)
	session.Values = map[any]any{"payment": paymentSession}
	session.Options = paymentCookieOptions
	return store.Save(r, w, session)
}

func ClearPayment(store Store, r *http.Request, w http.ResponseWriter) error {
	session, err := store.Get(r, cookiePay)
	if err != nil {
		return err
	}
	session.Values = map[any]any{}
	session.Options.MaxAge = -1
	return store.Save(r, w, session)
}

type ShareCodeSession struct {
	LpaID           string
	SessionID       string
	Identity        bool
	DonorFullName   string
	DonorFirstNames string
}

func (s ShareCodeSession) Valid() bool {
	return s.LpaID != ""
}

func ShareCode(store sessions.Store, r *http.Request) (*ShareCodeSession, error) {
	params, err := store.Get(r, cookieShareCode)
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["share-code"]
	if !ok {
		return nil, MissingSessionError("share-code")
	}

	shareCodeSession, ok := session.(*ShareCodeSession)
	if !ok {
		return nil, MissingSessionError("share-code")
	}
	if !shareCodeSession.Valid() {
		return nil, InvalidSessionError("share-code")
	}

	return shareCodeSession, nil
}

func SetShareCode(store sessions.Store, r *http.Request, w http.ResponseWriter, shareCodeSession *ShareCodeSession) error {
	session := sessions.NewSession(store, cookieShareCode)
	session.Values = map[any]any{"share-code": shareCodeSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}
