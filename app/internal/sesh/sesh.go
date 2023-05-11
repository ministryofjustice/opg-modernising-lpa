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
	gob.Register(&YotiSession{})
	gob.Register(&DonorSession{})
	gob.Register(&CertificateProviderSession{})
	gob.Register(&AttorneySession{})
	gob.Register(&PaymentSession{})
	gob.Register(&ShareCodeSession{})
}

type OneLoginSession struct {
	State               string
	Nonce               string
	Locale              string
	Identity            bool
	CertificateProvider bool
	SessionID           string
	LpaID               string
	Attorney            bool
}

func (s OneLoginSession) Valid() bool {
	ok := s.State != "" && s.Nonce != ""
	if s.CertificateProvider && !s.Identity {
		ok = ok && s.LpaID != ""
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

type YotiSession struct {
	Locale              string
	LpaID               string
	CertificateProvider bool
}

func (s YotiSession) Valid() bool {
	return s.LpaID != ""
}

func Yoti(store sessions.Store, r *http.Request) (*YotiSession, error) {
	params, err := store.Get(r, "yoti")
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
	params := sessions.NewSession(store, "yoti")
	params.Values = map[any]any{"yoti": yotiSession}
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
	Sub   string
	Email string
	LpaID string
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

type AttorneySession struct {
	Sub                   string
	Email                 string
	LpaID                 string
	DonorSessionID        string
	AttorneyID            string
	IsReplacementAttorney bool
}

func (s AttorneySession) Valid() bool {
	return s.Sub != ""
}

func Attorney(store sessions.Store, r *http.Request) (*AttorneySession, error) {
	params, err := store.Get(r, "session")
	if err != nil {
		return nil, err
	}

	session, ok := params.Values["attorney"]
	if !ok {
		return nil, MissingSessionError("attorney")
	}

	attorneySession, ok := session.(*AttorneySession)
	if !ok {
		return nil, MissingSessionError("attorney")
	}
	if !attorneySession.Valid() {
		return nil, InvalidSessionError("attorney")
	}

	return attorneySession, nil
}

func SetAttorney(store sessions.Store, r *http.Request, w http.ResponseWriter, attorneySession *AttorneySession) error {
	session := sessions.NewSession(store, "session")
	session.Values = map[any]any{"attorney": attorneySession}
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

type ShareCodeSession struct {
	LpaID           string
	Identity        bool
	DonorFullName   string
	DonorFirstNames string
}

func (s ShareCodeSession) Valid() bool {
	return s.LpaID != ""
}

func SetShareCode(store sessions.Store, r *http.Request, w http.ResponseWriter, shareCodeSession *ShareCodeSession) error {
	session := sessions.NewSession(store, "shareCode")
	session.Values = map[any]any{"share-code": shareCodeSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}

func ShareCode(store sessions.Store, r *http.Request) (*ShareCodeSession, error) {
	params, err := store.Get(r, "shareCode")
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
