package certificateprovider

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
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
	Create(ctx context.Context, sessionID string) (*actor.CertificateProviderProvidedDetails, error)
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

	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumber,
		EnterReferenceNumber(tmpls.Get("certificate_provider_enter_reference_number.gohtml"), shareCodeStore, sessionStore))
	handleRoot(page.Paths.CertificateProvider.WhoIsEligible,
		WhoIsEligible(tmpls.Get("certificate_provider_who_is_eligible.gohtml"), sessionStore))
	handleRoot(page.Paths.CertificateProvider.Login,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProvider.LoginCallback,
		LoginCallback(oneLoginClient, sessionStore, certificateProviderStore))

	certificateProviderMux := http.NewServeMux()
	rootMux.Handle("/certificate-provider/", page.RouteToPrefix("/certificate-provider/", certificateProviderMux, notFoundHandler))
	handleCertificateProvider := makeCertificateProviderHandle(certificateProviderMux, sessionStore, errorHandler)

	handleCertificateProvider(page.Paths.CertificateProvider.EnterDateOfBirth,
		EnterDateOfBirth(tmpls.Get("certificate_provider_enter_date_of_birth.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.ConfirmYourDetails,
		Guidance(tmpls.Get("certificate_provider_confirm_your_details.gohtml"), donorStore, certificateProviderStore))

	handleCertificateProvider(page.Paths.CertificateProvider.WhatYoullNeedToConfirmYourIdentity,
		Guidance(tmpls.Get("certificate_provider_what_youll_need_to_confirm_your_identity.gohtml"), donorStore, nil))

	for path, page := range map[page.CertificateProviderPath]int{
		page.Paths.CertificateProvider.SelectYourIdentityOptions:  0,
		page.Paths.CertificateProvider.SelectYourIdentityOptions1: 1,
		page.Paths.CertificateProvider.SelectYourIdentityOptions2: 2,
	} {
		handleCertificateProvider(path,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), page, certificateProviderStore))
	}

	handleCertificateProvider(page.Paths.CertificateProvider.YourChosenIdentityOptions,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithYoti,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), sessionStore, yotiClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithYotiCallback,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLogin,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLoginCallback,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, certificateProviderStore))

	for path, identityOption := range map[page.CertificateProviderPath]identity.Option{
		page.Paths.CertificateProvider.IdentityWithPassport:                 identity.Passport,
		page.Paths.CertificateProvider.IdentityWithBiometricResidencePermit: identity.BiometricResidencePermit,
		page.Paths.CertificateProvider.IdentityWithDrivingLicencePaper:      identity.DrivingLicencePaper,
		page.Paths.CertificateProvider.IdentityWithDrivingLicencePhotocard:  identity.DrivingLicencePhotocard,
		page.Paths.CertificateProvider.IdentityWithOnlineBankAccount:        identity.OnlineBankAccount,
	} {
		handleCertificateProvider(path,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), time.Now, identityOption, certificateProviderStore))
	}

	handleCertificateProvider(page.Paths.CertificateProvider.ReadTheLpa,
		Guidance(tmpls.Get("certificate_provider_read_the_lpa.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatHappensNext,
		Guidance(tmpls.Get("certificate_provider_what_happens_next.gohtml"), donorStore, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.ProvideCertificate,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), donorStore, time.Now, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.CertificateProvided,
		Guidance(tmpls.Get("certificate_provided.gohtml"), donorStore, nil))
}

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.Path, page.Handler) {
	return func(path page.Path, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.Page = path.Format()

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeCertificateProviderHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.CertificateProviderPath, page.Handler) {
	return func(path page.CertificateProviderPath, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider

			session, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
				return
			}

			appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				ctx = page.ContextWithSessionData(ctx, sessionData)

				appData.LpaID = sessionData.LpaID
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			appData.Page = path.Format(appData.LpaID)

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
