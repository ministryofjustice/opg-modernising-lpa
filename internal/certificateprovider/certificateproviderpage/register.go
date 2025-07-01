// Package certificateproviderpage provides the pages that a certificate
// provider interacts with.
package certificateproviderpage

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request, details *certificateproviderdata.Provided, lpa *lpadata.Lpa) error

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpadata.Lpa, error)
}

type EventClient interface {
	SendIdentityCheckMismatched(ctx context.Context, e event.IdentityCheckMismatched) error
	SendMetric(ctx context.Context, category event.Category, measure event.Measure) error
}

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type CertificateProviderStore interface {
	Create(ctx context.Context, link accesscodedata.Link, email string) (*certificateproviderdata.Provided, error)
	Delete(ctx context.Context) error
	Get(ctx context.Context) (*certificateproviderdata.Provided, error)
	Put(ctx context.Context, certificateProvider *certificateproviderdata.Provided) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, confidenceLevel onelogin.ConfidenceLevel) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(userInfo onelogin.UserInfo) (identity.UserData, error)
}

type AccessCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, code accesscodedata.Hashed) (accesscodedata.Link, error)
	Put(ctx context.Context, actorType actor.Type, code accesscodedata.Hashed, link accesscodedata.Link) error
	Delete(ctx context.Context, link accesscodedata.Link) error
}

type Template func(io.Writer, interface{}) error

type SessionStore interface {
	ClearLogin(r *http.Request, w http.ResponseWriter) error
	Login(r *http.Request) (*sesh.LoginSession, error)
	LpaData(r *http.Request) (*sesh.LpaDataSession, error)
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *sesh.LpaDataSession) error
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type NotifyClient interface {
	EmailGreeting(lpa *lpadata.Lpa) string
	SendActorEmail(ctx context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
}

type AccessCodeSender interface {
	SendAttorneys(context.Context, appcontext.Data, *lpadata.Lpa) error
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type Localizer interface {
	localize.Localizer
}

type DashboardStore interface {
	GetAll(ctx context.Context) (results dashboarddata.Results, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	Lpa(ctx context.Context, lpaUID string) (*lpadata.Lpa, error)
	SendCertificateProvider(ctx context.Context, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error
	SendCertificateProviderConfirmIdentity(ctx context.Context, lpaUID string, certificateProvider *certificateproviderdata.Provided) error
	SendCertificateProviderOptOut(ctx context.Context, lpaUID string, actorUID actoruid.UID) error
	SendPaperCertificateProviderAccessOnline(ctx context.Context, lpa *lpadata.Lpa, certificateProviderEmail string) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type DonorStore interface {
	GetAny(ctx context.Context) (*donordata.Provided, error)
	Put(ctx context.Context, donor *donordata.Provided) error
}

type ScheduledStore interface {
	Create(ctx context.Context, rows ...scheduled.Event) error
}

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
	sessionStore SessionStore,
	oneLoginClient OneLoginClient,
	accessCodeStore AccessCodeStore,
	errorHandler page.ErrorHandler,
	certificateProviderStore CertificateProviderStore,
	addressClient AddressClient,
	notifyClient NotifyClient,
	accessCodeSender AccessCodeSender,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
	lpaStoreResolvingService LpaStoreResolvingService,
	donorStore DonorStore,
	eventClient EventClient,
	scheduledStore ScheduledStore,
	bundle Bundle,
	donorStartURL string,
	certificateProviderStartURL string,
) {
	handleRoot := makeHandle(rootMux, errorHandler, sessionStore, certificateProviderStartURL)

	handleRoot(page.PathCertificateProviderLogin, page.None,
		page.Login(oneLoginClient, sessionStore, random.AlphaNumeric, page.PathCertificateProviderLoginCallback))
	handleRoot(page.PathCertificateProviderLoginCallback, page.None,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.PathCertificateProviderEnterAccessCode, dashboardStore, actor.TypeCertificateProvider))
	handleRoot(page.PathCertificateProviderEnterAccessCode, page.RequireSession,
		page.EnterAccessCode(tmpls.Get("enter_access_code.gohtml"), accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeCertificateProvider,
			EnterAccessCode(sessionStore, certificateProviderStore, lpaStoreClient, dashboardStore, eventClient)))
	handleRoot(page.PathCertificateProviderEnterAccessCodeOptOut, page.None,
		page.EnterAccessCodeOptOut(tmpls.Get("enter_access_code_opt_out.gohtml"), accessCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeCertificateProvider,
			page.PathCertificateProviderConfirmDontWantToBeCertificateProviderLoggedOut))
	handleRoot(page.PathCertificateProviderConfirmDontWantToBeCertificateProviderLoggedOut, page.None,
		ConfirmDontWantToBeCertificateProviderLoggedOut(tmpls.Get("confirm_dont_want_to_be_certificate_provider.gohtml"), accessCodeStore, lpaStoreResolvingService, lpaStoreClient, donorStore, sessionStore, notifyClient, donorStartURL))
	handleRoot(page.PathCertificateProviderYouHaveDecidedNotToBeCertificateProvider, page.None,
		page.Guidance(tmpls.Get("you_have_decided_not_to_be_a_certificate_provider.gohtml")))
	handleRoot(page.PathCertificateProviderYouHaveAlreadyProvidedACertificate, page.None,
		page.Guidance(tmpls.Get("you_have_already_provided_a_certificate_for_donors_lpa.gohtml")))
	handleRoot(page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn, page.RequireSession,
		page.Guidance(tmpls.Get("you_have_already_provided_a_certificate_for_donors_lpa.gohtml")))

	handleCertificateProvider := makeCertificateProviderHandle(rootMux, sessionStore, errorHandler, certificateProviderStore, lpaStoreResolvingService, certificateProviderStartURL)

	handleCertificateProvider(certificateprovider.PathWhoIsEligible, page.None,
		Guidance(tmpls.Get("who_is_eligible.gohtml")))
	handleCertificateProvider(certificateprovider.PathTaskList, page.None,
		TaskList(tmpls.Get("task_list.gohtml")))
	handleCertificateProvider(certificateprovider.PathEnterDateOfBirth, page.CanGoBack,
		EnterDateOfBirth(tmpls.Get("enter_date_of_birth.gohtml"), certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathYourPreferredLanguage, page.CanGoBack,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathWhatIsYourHomeAddress, page.None,
		WhatIsYourHomeAddress(logger, tmpls.Get("what_is_your_home_address.gohtml"), addressClient, certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathConfirmYourDetails, page.None,
		ConfirmYourDetails(tmpls.Get("confirm_your_details.gohtml"), certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathYourRole, page.CanGoBack,
		Guidance(tmpls.Get("your_role.gohtml")))
	handleCertificateProvider(certificateprovider.PathReadTheDraftLpa, page.None,
		Guidance(tmpls.Get("read_the_draft_lpa.gohtml")))

	handleCertificateProvider(certificateprovider.PathConfirmYourIdentity, page.None,
		ConfirmYourIdentity(tmpls.Get("confirm_your_identity.gohtml"), certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathHowWillYouConfirmYourIdentity, page.None,
		HowWillYouConfirmYourIdentity(tmpls.Get("how_will_you_confirm_your_identity.gohtml"), certificateProviderStore))
	handleCertificateProvider(certificateprovider.PathCompletingYourIdentityConfirmation, page.None,
		CompletingYourIdentityConfirmation(tmpls.Get("completing_your_identity_confirmation.gohtml")))
	handleCertificateProvider(certificateprovider.PathIdentityWithOneLogin, page.None,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.AlphaNumeric))
	handleCertificateProvider(certificateprovider.PathIdentityWithOneLoginCallback, page.None,
		IdentityWithOneLoginCallback(oneLoginClient, sessionStore, certificateProviderStore, lpaStoreClient, eventClient))
	handleCertificateProvider(certificateprovider.PathIdentityDetails, page.None,
		Guidance(tmpls.Get("identity_details.gohtml")))

	handleCertificateProvider(certificateprovider.PathReadTheLpa, page.None,
		ReadTheLpa(tmpls.Get("read_the_lpa.gohtml"), certificateProviderStore, bundle))
	handleCertificateProvider(certificateprovider.PathWhatHappensNext, page.CanGoBack,
		Guidance(tmpls.Get("what_happens_next.gohtml")))
	handleCertificateProvider(certificateprovider.PathProvideCertificate, page.CanGoBack,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), certificateProviderStore, notifyClient, accessCodeSender, lpaStoreClient, scheduledStore, donorStore, time.Now, donorStartURL))
	handleCertificateProvider(certificateprovider.PathCertificateProvided, page.None,
		Guidance(tmpls.Get("certificate_provided.gohtml")))
	handleCertificateProvider(certificateprovider.PathConfirmDontWantToBeCertificateProvider, page.CanGoBack,
		ConfirmDontWantToBeCertificateProvider(tmpls.Get("confirm_dont_want_to_be_certificate_provider.gohtml"), lpaStoreClient, donorStore, certificateProviderStore, notifyClient, donorStartURL))
}

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler, sessionStore SessionStore, certificateProviderStartURL string) func(page.Path, page.HandleOpt, page.Handler) {
	return func(path page.Path, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)

			if opt&page.RequireSession != 0 {
				loginSession, err := sessionStore.Login(r)
				if err != nil {
					http.Redirect(w, r, certificateProviderStartURL, http.StatusFound)
					return
				}

				appData.SessionID = loginSession.SessionID()
				appData.HasLpas = loginSession.HasLPAs

				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID})
			}

			appData.ActorType = actor.TypeCertificateProvider
			appData.Page = path.Format()

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeCertificateProviderHandle(mux *http.ServeMux, sessionStore SessionStore, errorHandler page.ErrorHandler, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStartURL string) func(certificateprovider.Path, page.HandleOpt, Handler) {
	return func(path certificateprovider.Path, opt page.HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.ActorType = actor.TypeCertificateProvider
			appData.CanGoBack = opt&page.CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := sessionStore.Login(r)
			if err != nil {
				http.Redirect(w, r, certificateProviderStartURL, http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()
			appData.HasLpas = session.HasLPAs

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

			lpa, err := lpaStoreResolvingService.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !path.CanGoTo(provided, lpa) {
				certificateprovider.PathTaskList.Redirect(w, r, appData, provided.LpaID)
				return
			}

			appData.Page = path.Format(appData.LpaID)

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData)), provided, lpa); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
