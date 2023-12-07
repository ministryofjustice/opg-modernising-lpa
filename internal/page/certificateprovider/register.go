package certificateprovider

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	GetAny(context.Context) (*actor.DonorProvidedDetails, error)
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
	Put(context.Context, actor.Type, string, actor.ShareCodeData) error
}

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	SendEmail(context.Context, string, notify.Email) (string, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeSender --structname mockShareCodeSender
type ShareCodeSender interface {
	SendAttorneys(context.Context, page.AppData, *actor.DonorProvidedDetails) error
}

//go:generate mockery --testonly --inpackage --name AddressClient --structname mockAddressClient
type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

//go:generate mockery --testonly --inpackage --name Localizer --structname mockLocalizer
type Localizer interface {
	Format(string, map[string]any) string
	T(string) string
	Count(string, int) string
	FormatCount(string, int, map[string]interface{}) string
	ShowTranslationKeys() bool
	SetShowTranslationKeys(bool)
	Possessive(string) string
	Concat([]string, string) string
	FormatDate(date.TimeOrDate) string
	FormatDateTime(time.Time) string
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
	certificateProviderStore CertificateProviderStore,
	notFoundHandler page.Handler,
	addressClient AddressClient,
	notifyClient NotifyClient,
	shareCodeSender ShareCodeSender,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.CertificateProvider.Login,
		page.Login(logger, oneLoginClient, sessionStore, random.String, page.Paths.CertificateProvider.LoginCallback))
	handleRoot(page.Paths.CertificateProvider.LoginCallback,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.CertificateProvider.EnterReferenceNumber))
	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumber,
		EnterReferenceNumber(tmpls.Get("certificate_provider_enter_reference_number.gohtml"), shareCodeStore, sessionStore, certificateProviderStore))

	certificateProviderMux := http.NewServeMux()
	rootMux.Handle("/certificate-provider/", page.RouteToPrefix("/certificate-provider/", certificateProviderMux, notFoundHandler))
	handleCertificateProvider := makeCertificateProviderHandle(certificateProviderMux, sessionStore, errorHandler)

	handleCertificateProvider(page.Paths.CertificateProvider.WhoIsEligible, page.None,
		WhoIsEligible(tmpls.Get("certificate_provider_who_is_eligible.gohtml"), donorStore))
	handleCertificateProvider(page.Paths.CertificateProvider.TaskList, page.None,
		TaskList(tmpls.Get("certificate_provider_task_list.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.EnterDateOfBirth, page.CanGoBack,
		EnterDateOfBirth(tmpls.Get("certificate_provider_enter_date_of_birth.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.YourPreferredLanguage, page.CanGoBack,
		YourPreferredLanguage(tmpls.Get("your_preferred_language.gohtml"), certificateProviderStore, donorStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatIsYourHomeAddress, page.None,
		WhatIsYourHomeAddress(logger, tmpls.Get("certificate_provider_what_is_your_home_address.gohtml"), addressClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.ConfirmYourDetails, page.None,
		ConfirmYourDetails(tmpls.Get("certificate_provider_confirm_your_details.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.YourRole, page.CanGoBack,
		Guidance(tmpls.Get("certificate_provider_your_role.gohtml"), donorStore, nil))

	handleCertificateProvider(page.Paths.CertificateProvider.ProveYourIdentity, page.None,
		Guidance(tmpls.Get("certificate_provider_prove_your_identity.gohtml"), nil, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLogin, page.None,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLoginCallback, page.None,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, certificateProviderStore, donorStore))

	handleCertificateProvider(page.Paths.CertificateProvider.ReadTheLpa, page.None,
		ReadTheLpa(tmpls.Get("certificate_provider_read_the_lpa.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatHappensNext, page.CanGoBack,
		Guidance(tmpls.Get("certificate_provider_what_happens_next.gohtml"), donorStore, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.ProvideCertificate, page.CanGoBack,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), donorStore, time.Now, certificateProviderStore, notifyClient, shareCodeSender))
	handleCertificateProvider(page.Paths.CertificateProvider.CertificateProvided, page.None,
		Guidance(tmpls.Get("certificate_provided.gohtml"), donorStore, certificateProviderStore))
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

func makeCertificateProviderHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.CertificateProviderPath, page.HandleOpt, page.Handler) {
	return func(path page.CertificateProviderPath, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.CanGoBack = opt&page.CanGoBack != 0

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
