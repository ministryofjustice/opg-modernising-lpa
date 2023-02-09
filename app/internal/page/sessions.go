package page

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

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
		Secure:   true, // check this works
	}
	oneLoginCookieOptions = &sessions.Options{
		Path:     "/",
		MaxAge:   10 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
)

func init() {
	gob.Register(&OneLoginSession{})
	gob.Register(&DonorSession{})
	gob.Register(&CertificateProviderSession{})
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

func getOneLoginSession(store sessions.Store, r *http.Request) (*OneLoginSession, error) {
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

func setOneLoginSession(store sessions.Store, r *http.Request, w http.ResponseWriter, oneLoginSession *OneLoginSession) error {
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

func getDonorSession(store sessions.Store, r *http.Request) (*DonorSession, error) {
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

func setDonorSession(store sessions.Store, r *http.Request, w http.ResponseWriter, donorSession *DonorSession) error {
	session := sessions.NewSession(store, "session")
	session.Values = map[any]any{"donor": donorSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}

type CertificateProviderSession struct {
	Sub       string
	Email     string
	LpaID     string
	SessionID string // this is the donor's sessionID
}

func (s CertificateProviderSession) Valid() bool {
	return s.Sub != ""
}

func getCertificateProviderSession(store sessions.Store, r *http.Request) (*CertificateProviderSession, error) {
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

func setCertificateProviderSession(store sessions.Store, r *http.Request, w http.ResponseWriter, certificateProviderSession *CertificateProviderSession) error {
	session := sessions.NewSession(store, "session")
	session.Values = map[any]any{"certificate-provider": certificateProviderSession}
	session.Options = sessionCookieOptions
	return store.Save(r, w, session)
}
