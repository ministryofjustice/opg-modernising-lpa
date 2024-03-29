package certificateprovider

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type DonorStore interface {
	GetAny(context.Context) (*actor.DonorProvidedDetails, error)
}

type CertificateProviderStore interface {
	Create(ctx context.Context, sessionID string, certificateProviderUID actoruid.UID) (*actor.CertificateProviderProvidedDetails, error)
	Get(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
	Put(ctx context.Context, certificateProvider *actor.CertificateProviderProvidedDetails) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (actor.ShareCodeData, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, shareCodeData actor.ShareCodeData) error
	Delete(ctx context.Context, shareCode actor.ShareCodeData) error
}

type Template func(io.Writer, interface{}) error

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type NotifyClient interface {
	SendEmail(ctx context.Context, to string, email notify.Email) error
	SendActorEmail(ctx context.Context, to, lpaUID string, email notify.Email) error
}

type ShareCodeSender interface {
	SendAttorneys(context.Context, page.AppData, *actor.DonorProvidedDetails) error
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type Localizer interface {
	page.Localizer
}

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	SendCertificateProvider(context.Context, string, *actor.CertificateProviderProvidedDetails) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
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
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
) {
	handleRoot := makeHandle(rootMux, errorHandler)

	handleRoot(page.Paths.CertificateProvider.Login,
		page.Login(oneLoginClient, sessionStore, random.String, page.Paths.CertificateProvider.LoginCallback))
	handleRoot(page.Paths.CertificateProvider.LoginCallback,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.CertificateProvider.EnterReferenceNumber, dashboardStore, actor.TypeCertificateProvider))
	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumber,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, certificateProviderStore))

	handleCertificateProvider := makeCertificateProviderHandle(rootMux, sessionStore, errorHandler)

	handleCertificateProvider(page.Paths.CertificateProvider.WhoIsEligible, page.None,
		WhoIsEligible(tmpls.Get("who_is_eligible.gohtml"), donorStore))
	handleCertificateProvider(page.Paths.CertificateProvider.TaskList, page.None,
		TaskList(tmpls.Get("task_list.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.EnterDateOfBirth, page.CanGoBack,
		EnterDateOfBirth(tmpls.Get("enter_date_of_birth.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.YourPreferredLanguage, page.CanGoBack,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), certificateProviderStore, donorStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatIsYourHomeAddress, page.None,
		WhatIsYourHomeAddress(logger, tmpls.Get("what_is_your_home_address.gohtml"), addressClient, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.ConfirmYourDetails, page.None,
		ConfirmYourDetails(tmpls.Get("confirm_your_details.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.YourRole, page.CanGoBack,
		Guidance(tmpls.Get("your_role.gohtml"), donorStore, nil))

	handleCertificateProvider(page.Paths.CertificateProvider.ProveYourIdentity, page.None,
		Guidance(tmpls.Get("prove_your_identity.gohtml"), nil, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLogin, page.None,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.String))
	handleCertificateProvider(page.Paths.CertificateProvider.IdentityWithOneLoginCallback, page.None,
		IdentityWithOneLoginCallback(commonTmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, certificateProviderStore, donorStore))

	handleCertificateProvider(page.Paths.CertificateProvider.ReadTheLpa, page.None,
		ReadTheLpa(tmpls.Get("read_the_lpa.gohtml"), donorStore, certificateProviderStore))
	handleCertificateProvider(page.Paths.CertificateProvider.WhatHappensNext, page.CanGoBack,
		Guidance(tmpls.Get("what_happens_next.gohtml"), donorStore, nil))
	handleCertificateProvider(page.Paths.CertificateProvider.ProvideCertificate, page.CanGoBack,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), donorStore, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, time.Now))
	handleCertificateProvider(page.Paths.CertificateProvider.CertificateProvided, page.None,
		Guidance(tmpls.Get("certificate_provided.gohtml"), donorStore, certificateProviderStore))
}

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler) func(page.Path, page.Handler) {
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

func makeCertificateProviderHandle(mux *http.ServeMux, sessionStore SessionStore, errorHandler page.ErrorHandler) func(page.CertificateProviderPath, page.HandleOpt, page.Handler) {
	return func(path page.CertificateProviderPath, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.CanGoBack = opt&page.CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := sessionStore.Login(r)
			if err != nil {
				http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = page.ContextWithSessionData(ctx, sessionData)
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
