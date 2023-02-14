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
		MaxAge:   10 * 60,
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
	gob.Register(&DonorSession{})
	gob.Register(&CertificateProviderSession{})
	gob.Register(&PaymentSession{})
}

type OneLoginSession struct {
	State               string
	Nonce               string
	Locale              string
	Identity            bool
	CertificateProvider bool
	SessionID           string
	LpaID               string
}

func (s OneLoginSession) Valid() bool {
	ok := s.State != "" && s.Nonce != ""
	if s.CertificateProvider {
		ok = ok && s.SessionID != "" && s.LpaID != ""
	}

	return ok
}

func OneLogin(store sessions.Store, r *http.Request) (*OneLoginSession, error) {
	params, err := store.Get(r, "params")
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
	params := sessions.NewSession(store, "params")
	params.Values = map[any]any{"one-login": oneLoginSession}
	params.Options = oneLoginCookieOptions
	return store.Save(r, w, params)
}

type DonorSession struct {
	Sub   string
	Email string
}

func (s DonorSession) Valid() bool {
	return s.Sub != ""
}

func Donor(store sessions.Store, r *http.Request) (*DonorSession, error) {
	params, err := store.Get(r, "session")
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["donor"]
	if !ok {
		return nil, MissingSessionError("donor")
	}

	donorSession, ok := session.(*DonorSession)
	if !ok {
		return nil, MissingSessionError("donor")
	}
	if !donorSession.Valid() {
		return nil, InvalidSessionError("donor")
	}

	return donorSession, nil
}

func SetDonor(store sessions.Store, r *http.Request, w http.ResponseWriter, donorSession *DonorSession) error {
	session := sessions.NewSession(store, "session")
	session.Values = map[any]any{"donor": donorSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}

type CertificateProviderSession struct {
	Sub            string
	Email          string
	LpaID          string
	DonorSessionID string
}

func (s CertificateProviderSession) Valid() bool {
	return s.Sub != ""
}

func CertificateProvider(store sessions.Store, r *http.Request) (*CertificateProviderSession, error) {
	params, err := store.Get(r, "session")
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["certificate-provider"]
	if !ok {
		return nil, MissingSessionError("certificate-provider")
	}

	certificateProviderSession, ok := session.(*CertificateProviderSession)
	if !ok {
		return nil, MissingSessionError("certificate-provider")
	}
	if !certificateProviderSession.Valid() {
		return nil, InvalidSessionError("certificate-provider")
	}

	return certificateProviderSession, nil
}

func SetCertificateProvider(store sessions.Store, r *http.Request, w http.ResponseWriter, certificateProviderSession *CertificateProviderSession) error {
	session := sessions.NewSession(store, "session")
	session.Values = map[any]any{"certificate-provider": certificateProviderSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}

type PaymentSession struct {
	PaymentID string
}

func (s PaymentSession) Valid() bool {
	return true
}

func Payment(store sessions.Store, r *http.Request) (*PaymentSession, error) {
	params, err := store.Get(r, "pay")
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
	session := sessions.NewSession(store, "pay")
	session.Values = map[any]any{"payment": paymentSession}
	session.Options = paymentCookieOptions
	return store.Save(r, w, session)
}

func ClearPayment(store Store, r *http.Request, w http.ResponseWriter) error {
	session, err := store.Get(r, "pay")
	if err != nil {
		return err
	}
	session.Values = map[any]any{}
	session.Options.MaxAge = -1
	return store.Save(r, w, session)
}
