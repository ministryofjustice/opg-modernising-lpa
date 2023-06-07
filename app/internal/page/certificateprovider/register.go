package certificateprovider

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	GetAny(context.Context) (*page.Lpa, error)
}

//go:generate mockery --testonly --inpackage --name CertificateProviderStore --structname mockCertificateProviderStore
type CertificateProviderStore interface {
	Create(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
	Get(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
	Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeStore --structname mockShareCodeStore
type ShareCodeStore interface {
	Get(context.Context, actor.Type, string) (actor.ShareCodeData, error)
}

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name YotiClient --structname mockYotiClient
type YotiClient interface {
	IsTest() bool
	SdkID() string
	ScenarioID() string
	User(string) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	donorStore DonorStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	yotiClient YotiClient,
	notifyClient NotifyClient,
	certificateProviderStore CertificateProviderStore,
	notFoundHandler page.Handler,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumber, None,
		EnterReferenceNumber(tmpls.Get("certificate_provider_enter_reference_number.gohtml"), shareCodeStore, sessionStore))
	handleRoot(page.Paths.CertificateProvider.WhoIsEligible, None,
		WhoIsEligible(tmpls.Get("certificate_provider_who_is_eligible.gohtml"), sessionStore))
	handleRoot(page.Paths.CertificateProvider.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProvider.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore, certificateProviderStore))

	certificateProviderMux := http.NewServeMux()
	rootMux.Handle("/certificate-provider/", routeToPrefix("/certificate-provider/", certificateProviderMux, notFoundHandler))
	handleCertificateProvider := makeHandle(certificateProviderMux, sessionStore, errorHandler)

	handleCertificateProvider(page.Paths.CertificateProvider.CheckYourName, RequireSession,
		CheckYourName(tmpls.Get("certificate_provider_check_your_name.gohtml"), donorStore, notifyClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.EnterDateOfBirth, RequireSession,
		EnterDateOfBirth(tmpls.Get("certificate_provider_enter_date_of_birth.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.EnterMobileNumber, RequireSession,
		EnterMobileNumber(tmpls.Get("certificate_provider_enter_mobile_number.gohtml"), donorStore, certificateProviderStore))

	handleCertificateProvider(page.Paths.CertificateProvider.WhatYoullNeedToConfirmYourIdentity, RequireSession,
		Guidance(tmpls.Get("certificate_provider_what_youll_need_to_confirm_your_identity.gohtml"), donorStore, nil))

	for path, page := range map[string]int{
		page.Paths.CertificateProvider.SelectYourIdentityOptions:  0,
		page.Paths.CertificateProvider.SelectYourIdentityOptions1: 1,
		page.Paths.CertificateProvider.SelectYourIdentityOptions2: 2,
	} {
		handleCertificateProvider(path, RequireSession,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), page, certificateProviderStore))
	}

	handleCertificateProvider(page.Paths.CertificateProvider.YourChosenIdentityOptions, RequireSession,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithYoti, RequireSession,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), sessionStore, yotiClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithYotiCallback, RequireSession,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLogin, RequireSession,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLoginCallback, RequireSession,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, certificateProviderStore))

	for path, identityOption := range map[string]identity.Option{
		page.Paths.CertificateProvider.IdentityWithPassport:                 identity.Passport,
		page.Paths.CertificateProvider.IdentityWithBiometricResidencePermit: identity.BiometricResidencePermit,
		page.Paths.CertificateProvider.IdentityWithDrivingLicencePaper:      identity.DrivingLicencePaper,
		page.Paths.CertificateProvider.IdentityWithDrivingLicencePhotocard:  identity.DrivingLicencePhotocard,
		page.Paths.CertificateProvider.IdentityWithOnlineBankAccount:        identity.OnlineBankAccount,
	} {
		handleCertificateProvider(path, RequireSession,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), time.Now, identityOption, certificateProviderStore))
	}

	handleCertificateProvider(page.Paths.CertificateProvider.ReadTheLpa, RequireSession,
		Guidance(tmpls.Get("certificate_provider_read_the_lpa.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatHappensNext, RequireSession,
		Guidance(tmpls.Get("certificate_provider_what_happens_next.gohtml"), donorStore, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.ProvideCertificate, RequireSession,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), donorStore, time.Now, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.CertificateProvided, RequireSession,
		Guidance(tmpls.Get("certificate_provided.gohtml"), donorStore, nil))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ServiceName = "beACertificateProvider"
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0
			appData.ActorType = actor.TypeCertificateProvider

			if opt&RequireSession != 0 {
				session, err := sesh.CertificateProvider(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.CertificateProviderStart, http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))
				appData.LpaID = session.LpaID

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{
					SessionID: appData.SessionID,
					LpaID:     appData.LpaID,
				})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func routeToPrefix(prefix string, mux http.Handler, notFoundHandler page.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 4)
		if len(parts) != 4 {
			notFoundHandler(page.AppDataFromContext(r.Context()), w, r)
			return
		}

		id, path := parts[2], "/"+parts[3]

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		if len(r.URL.RawPath) > len(prefix)+len(id) {
			r2.URL.RawPath = r.URL.RawPath[len(prefix)+len(id):]
		}

		mux.ServeHTTP(w, r2.WithContext(page.ContextWithSessionData(r2.Context(), &page.SessionData{LpaID: id})))
	}
}
