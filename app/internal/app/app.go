package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	Get(ctx context.Context, pk, sk string, v interface{}) error
	Put(context.Context, string, string, interface{}) error
	GetOneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
	GetAllByGsi(ctx context.Context, gsi, sk string, v interface{}) error
	GetAllByKeys(ctx context.Context, pks []dynamo.Key) ([]map[string]types.AttributeValue, error)
	Create(ctx context.Context, pk, sk string, v interface{}) error
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

func App(
	logger *logging.Logger,
	localizer page.Localizer,
	lang localize.Lang,
	tmpls template.Templates,
	sessionStore SessionStore,
	dataStore DataStore,
	appPublicURL string,
	payClient *pay.Client,
	yotiClient *identity.YotiClient,
	notifyClient *notify.Client,
	addressClient *place.Client,
	rumConfig page.RumConfig,
	staticHash string,
	paths page.AppPaths,
	oneLoginClient *onelogin.Client,
	uidClient *uid.Client,
) http.Handler {
	donorStore := &donorStore{dataStore: dataStore, uuidString: uuid.NewString, now: time.Now}
	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: time.Now}
	attorneyStore := &attorneyStore{dataStore: dataStore, now: time.Now}
	shareCodeStore := &shareCodeStore{dataStore: dataStore}
	dashboardStore := &dashboardStore{dataStore: dataStore}

	shareCodeSender := page.NewShareCodeSender(shareCodeStore, notifyClient, appPublicURL, random.String)
	witnessCodeSender := page.NewWitnessCodeSender(donorStore, notifyClient)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger)

	rootMux := http.NewServeMux()

	rootMux.Handle(paths.TestingStart, page.TestingStart(sessionStore, donorStore, random.String, shareCodeSender, localizer, certificateProviderStore, attorneyStore, logger, time.Now))

	handleRoot := makeHandle(rootMux, errorHandler, sessionStore)

	handleRoot(paths.Root, None,
		notFoundHandler)
	handleRoot(paths.SignOut, None,
		page.SignOut(logger, sessionStore, oneLoginClient, appPublicURL))
	handleRoot(paths.Fixtures, None,
		page.Fixtures(tmpls.Get("fixtures.gohtml")))
	handleRoot(paths.YourLegalRightsAndResponsibilities, None,
		page.Guidance(tmpls.Get("your_legal_rights_and_responsibilities_general.gohtml")))
	handleRoot(page.Paths.Start, None,
		page.Guidance(tmpls.Get("start.gohtml")))
	handleRoot(page.Paths.CertificateProviderStart, None,
		page.Guidance(tmpls.Get("certificate_provider_start.gohtml")))
	handleRoot(page.Paths.Attorney.Start, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml")))
	handleRoot(page.Paths.Dashboard, RequireSession,
		page.Dashboard(tmpls.Get("dashboard.gohtml"), donorStore, dashboardStore))

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		donorStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		yotiClient,
		notifyClient,
		certificateProviderStore,
		notFoundHandler,
	)

	attorney.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		donorStore,
		certificateProviderStore,
		attorneyStore,
		oneLoginClient,
		addressClient,
		shareCodeStore,
		errorHandler,
		notifyClient,
		notFoundHandler,
	)

	donor.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		donorStore,
		oneLoginClient,
		addressClient,
		appPublicURL,
		payClient,
		yotiClient,
		notifyClient,
		shareCodeSender,
		witnessCodeSender,
		errorHandler,
		notFoundHandler,
		certificateProviderStore,
		uidClient,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.String, errorHandler), localizer, lang, rumConfig, staticHash)
}

func withAppData(next http.Handler, localizer page.Localizer, lang localize.Lang, rumConfig page.RumConfig, staticHash string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		localizer.SetShowTranslationKeys(r.FormValue("showTranslationKeys") == "1")

		appData := page.AppDataFromContext(ctx)
		appData.Path = r.URL.Path
		appData.Query = queryString(r)
		appData.Localizer = localizer
		appData.Lang = lang
		appData.RumConfig = rumConfig
		appData.StaticHash = staticHash
		appData.Paths = page.Paths
		appData.ActorTypes = actor.ActorTypes

		_, cookieErr := r.Cookie("cookies-consent")
		appData.CookieConsentSet = cookieErr != http.ErrNoCookie

		next.ServeHTTP(w, r.WithContext(page.ContextWithAppData(ctx, appData)))
	}
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
)

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler, store sesh.Store) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path

			if opt&RequireSession != 0 {
				loginSession, err := sesh.Login(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start, http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(loginSession.Sub))
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func queryString(r *http.Request) string {
	if r.URL.RawQuery != "" {
		return fmt.Sprintf("?%s", r.URL.RawQuery)
	} else {
		return ""
	}
}
