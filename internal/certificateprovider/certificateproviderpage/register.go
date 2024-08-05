// Package certificateproviderpage provides the pages that a certificate
// provider interacts with.
package certificateproviderpage

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
)

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request, details *certificateproviderdata.Provided) error

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpastore.Lpa, error)
}

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type CertificateProviderStore interface {
	Create(ctx context.Context, shareCode sharecode.Data, email string) (*certificateproviderdata.Provided, error)
	Delete(ctx context.Context) error
	Get(ctx context.Context) (*certificateproviderdata.Provided, error)
	Put(ctx context.Context, certificateProvider *certificateproviderdata.Provided) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (sharecode.Data, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, shareCodeData sharecode.Data) error
	Delete(ctx context.Context, shareCode sharecode.Data) error
}

type Template func(io.Writer, interface{}) error

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
	LpaData(r *http.Request) (*sesh.LpaDataSession, error)
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *sesh.LpaDataSession) error
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type NotifyClient interface {
	SendEmail(ctx context.Context, to string, email notify.Email) error
	SendActorEmail(ctx context.Context, to, lpaUID string, email notify.Email) error
}

type ShareCodeSender interface {
	SendAttorneys(context.Context, appcontext.Data, *lpastore.Lpa) error
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
	SendCertificateProvider(ctx context.Context, certificateProvider *certificateproviderdata.Provided, lpa *lpastore.Lpa) error
	SendCertificateProviderConfirmIdentity(ctx context.Context, lpaUID string, certificateProvider *certificateproviderdata.Provided) error
	SendCertificateProviderOptOut(ctx context.Context, lpaUID string, actorUID actoruid.UID) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type DonorStore interface {
	GetAny(ctx context.Context) (*donordata.Provided, error)
	Put(ctx context.Context, donor *donordata.Provided) error
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
	sessionStore SessionStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	certificateProviderStore CertificateProviderStore,
	addressClient AddressClient,
	notifyClient NotifyClient,
	shareCodeSender ShareCodeSender,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
	lpaStoreResolvingService LpaStoreResolvingService,
	donorStore DonorStore,
	appPublicURL string,
) {
	handleRoot := makeHandle(rootMux, errorHandler)

	handleRoot(page.Paths.CertificateProvider.Login,
		page.Login(oneLoginClient, sessionStore, random.String, page.Paths.CertificateProvider.LoginCallback))
	handleRoot(page.Paths.CertificateProvider.LoginCallback,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.Paths.CertificateProvider.EnterReferenceNumber, dashboardStore, actor.TypeCertificateProvider))
	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumber,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, certificateProviderStore))
	handleRoot(page.Paths.CertificateProvider.EnterReferenceNumberOptOut,
		EnterReferenceNumberOptOut(tmpls.Get("enter_reference_number_opt_out.gohtml"), shareCodeStore, sessionStore))
	handleRoot(page.Paths.CertificateProvider.ConfirmDontWantToBeCertificateProviderLoggedOut,
		ConfirmDontWantToBeCertificateProviderLoggedOut(tmpls.Get("confirm_dont_want_to_be_certificate_provider.gohtml"), shareCodeStore, lpaStoreResolvingService, lpaStoreClient, donorStore, sessionStore, notifyClient, appPublicURL))
	handleRoot(page.Paths.CertificateProvider.YouHaveDecidedNotToBeCertificateProvider,
		page.Guidance(tmpls.Get("you_have_decided_not_to_be_a_certificate_provider.gohtml")))

	handleCertificateProvider := makeCertificateProviderHandle(rootMux, sessionStore, errorHandler, certificateProviderStore)

	handleCertificateProvider(certificateprovider.PathWhoIsEligible, page.None,
		Guidance(tmpls.Get("who_is_eligible.gohtml"), lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathTaskList, page.None,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathEnterDateOfBirth, page.CanGoBack,
		EnterDateOfBirth(tmpls.Get("enter_date_of_birth.gohtml"), lpaStoreResolvingService, certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathYourPreferredLanguage, page.CanGoBack,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), certificateProviderStore, lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathWhatIsYourHomeAddress, page.None,
		WhatIsYourHomeAddress(logger, tmpls.Get("what_is_your_home_address.gohtml"), addressClient, certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathConfirmYourDetails, page.None,
		ConfirmYourDetails(tmpls.Get("confirm_your_details.gohtml"), lpaStoreResolvingService, certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathYourRole, page.CanGoBack,
		Guidance(tmpls.Get("your_role.gohtml"), lpaStoreResolvingService))

	handleCertificateProvider(certificateprovider.PathProveYourIdentity, page.None,
		Guidance(tmpls.Get("prove_your_identity.gohtml"), nil))
	handleCertificateProvider(certificateprovider.PathIdentityWithOneLogin, page.None,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.String))
	handleCertificateProvider(certificateprovider.PathIdentityWithOneLoginCallback, page.None,
		IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreResolvingService, notifyClient, lpaStoreClient, appPublicURL))
	handleCertificateProvider(certificateprovider.PathOneLoginIdentityDetails, page.None,
		OneLoginIdentityDetails(tmpls.Get("onelogin_identity_details.gohtml"), certificateProviderStore, lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathUnableToConfirmIdentity, page.None,
		UnableToConfirmIdentity(tmpls.Get("unable_to_confirm_identity.gohtml"), certificateProviderStore, lpaStoreResolvingService))

	handleCertificateProvider(certificateprovider.PathReadTheLpa, page.None,
		ReadTheLpa(tmpls.Get("read_the_lpa.gohtml"), lpaStoreResolvingService, certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathWhatHappensNext, page.CanGoBack,
		Guidance(tmpls.Get("what_happens_next.gohtml"), lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathProvideCertificate, page.CanGoBack,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), lpaStoreResolvingService, certificateProviderStore, notifyClient, shareCodeSender, lpaStoreClient, time.Now))
	handleCertificateProvider(certificateprovider.PathCertificateProvided, page.None,
		Guidance(tmpls.Get("certificate_provided.gohtml"), lpaStoreResolvingService))
	handleCertificateProvider(certificateprovider.PathConfirmDontWantToBeCertificateProvider, page.CanGoBack,
		ConfirmDontWantToBeCertificateProvider(tmpls.Get("confirm_dont_want_to_be_certificate_provider.gohtml"), lpaStoreResolvingService, lpaStoreClient, donorStore, certificateProviderStore, notifyClient, appPublicURL))
}

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler) func(page.Path, page.Handler) {
	return func(path page.Path, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.Page = path.Format()

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeCertificateProviderHandle(mux *http.ServeMux, sessionStore SessionStore, errorHandler page.ErrorHandler, certificateProviderStore CertificateProviderStore) func(certificateprovider.Path, page.HandleOpt, Handler) {
	return func(path certificateprovider.Path, opt page.HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.CanGoBack = opt&page.CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := sessionStore.Login(r)
			if err != nil {
				http.Redirect(w, r, page.Paths.CertificateProviderStart.Format(), http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()

			sessionData, err := appcontext.SessionFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = appcontext.ContextWithSession(ctx, sessionData)
			} else {
				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			provided, err := certificateProviderStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !certificateprovider.CanGoTo(provided, r.URL.String()) {
				page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, provided.LpaID)
				return
			}

			appData.Page = path.Format(appData.LpaID)

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData)), provided); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
