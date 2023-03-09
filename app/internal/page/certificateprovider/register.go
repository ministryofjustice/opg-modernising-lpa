package certificateprovider

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
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

//go:generate mockery --testonly --inpackage --name LpaStore --structname mockLpaStore
type LpaStore interface {
	Create(context.Context) (*page.Lpa, error)
	GetAll(context.Context) ([]*page.Lpa, error)
	Get(context.Context) (*page.Lpa, error)
	Put(context.Context, *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	GetAll(context.Context, string, interface{}) error
	Get(context.Context, string, string, interface{}) error
	Put(context.Context, string, string, interface{}) error
}

//go:generate mockery --testonly --inpackage --name AddressClient --structname mockAddressClient
type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
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
	lpaStore LpaStore,
	oneLoginClient OneLoginClient,
	dataStore DataStore,
	addressClient AddressClient,
	errorHandler page.ErrorHandler,
	yotiClient YotiClient,
	yotiScenarioID string,
	notifyClient NotifyClient,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.CertificateProviderStart, None,
		page.Guidance(tmpls.Get("certificate_provider_start.gohtml"), nil))
	handleRoot(page.Paths.CertificateProviderEnterReferenceNumber, None,
		EnterReferenceNumber(tmpls.Get("certificate_provider_enter_reference_number.gohtml"), lpaStore, dataStore))
	handleRoot(page.Paths.CertificateProviderLogin, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProviderLoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.CertificateProviderCheckYourName, RequireSession,
		CheckYourName(tmpls.Get("certificate_provider_check_your_name.gohtml"), lpaStore, notifyClient))
	handleRoot(page.Paths.CertificateProviderYourDetails, RequireSession,
		YourDetails(tmpls.Get("certificate_provider_your_details.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderYourAddress, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))

	handleRoot(page.Paths.CertificateProviderWhatYoullNeedToConfirmYourIdentity, RequireSession,
		page.Guidance(tmpls.Get("what_youll_need_to_confirm_your_identity.gohtml"), lpaStore))

	for path, page := range map[string]int{
		page.Paths.CertificateProviderSelectYourIdentityOptions:  0,
		page.Paths.CertificateProviderSelectYourIdentityOptions1: 1,
		page.Paths.CertificateProviderSelectYourIdentityOptions2: 2,
	} {
		handleRoot(path, RequireSession,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), lpaStore, page))
	}

	handleRoot(page.Paths.CertificateProviderYourChosenIdentityOptions, RequireSession,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderIdentityWithYoti, RequireSession,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), lpaStore, yotiClient, yotiScenarioID))
	handleRoot(page.Paths.CertificateProviderIdentityWithYotiCallback, RequireSession,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, lpaStore))
	handleRoot(page.Paths.CertificateProviderIdentityWithOneLogin, RequireSession,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProviderIdentityWithOneLoginCallback, RequireSession,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))

	for path, identityOption := range map[string]identity.Option{
		page.Paths.CertificateProviderIdentityWithPassport:                 identity.Passport,
		page.Paths.CertificateProviderIdentityWithBiometricResidencePermit: identity.BiometricResidencePermit,
		page.Paths.CertificateProviderIdentityWithDrivingLicencePaper:      identity.DrivingLicencePaper,
		page.Paths.CertificateProviderIdentityWithDrivingLicencePhotocard:  identity.DrivingLicencePhotocard,
		page.Paths.CertificateProviderIdentityWithOnlineBankAccount:        identity.OnlineBankAccount,
	} {
		handleRoot(path, RequireSession,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), lpaStore, time.Now, identityOption))
	}

	handleRoot(page.Paths.CertificateProviderReadTheLpa, RequireSession,
		page.Guidance(tmpls.Get("certificate_provider_read_the_lpa.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderWhatHappensNext, RequireSession,
		page.Guidance(tmpls.Get("certificate_provider_what_happens_next.gohtml"), lpaStore))
	handleRoot(page.Paths.ProvideCertificate, RequireSession,
		ProvideCertificate(tmpls.Get("provide_certificate.gohtml"), lpaStore, time.Now))
	handleRoot(page.Paths.CertificateProvided, RequireSession,
		page.Guidance(tmpls.Get("certificate_provided.gohtml"), lpaStore))
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
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := sesh.CertificateProvider(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.CertificateProviderStart, http.StatusFound)
					return
				}

				appData.SessionID = session.DonorSessionID
				appData.LpaID = session.LpaID

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
