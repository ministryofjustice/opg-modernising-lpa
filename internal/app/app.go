package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/fixtures"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name DynamoClient --structname mockDynamoClient
type DynamoClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
	AllByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
	LatestForActor(ctx context.Context, sk string, v interface{}) error
	AllForActor(ctx context.Context, sk string, v interface{}) error
	AllByKeys(ctx context.Context, pks []dynamo.Key) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllKeysByPk(ctx context.Context, pk string) ([]dynamo.Key, error)
	Put(ctx context.Context, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Key) error
	DeleteOne(ctx context.Context, pk, sk string) error
	Update(ctx context.Context, pk, sk string, values map[string]dynamodbtypes.AttributeValue, expression string) error
	BatchPut(ctx context.Context, items []interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
}

//go:generate mockery --testonly --inpackage --name S3Client --structname mockS3Client
type S3Client interface {
	PutObject(context.Context, string, []byte) error
	DeleteObject(context.Context, string) error
	DeleteObjects(ctx context.Context, keys []string) error
	PutObjectTagging(context.Context, string, map[string]string) error
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
	lpaDynamoClient DynamoClient,
	appPublicURL string,
	payClient *pay.Client,
	notifyClient *notify.Client,
	addressClient *place.Client,
	rumConfig page.RumConfig,
	staticHash string,
	paths page.AppPaths,
	oneLoginClient *onelogin.Client,
	oneloginURL string,
	s3Client S3Client,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
) http.Handler {
	documentStore := NewDocumentStore(lpaDynamoClient, s3Client, eventClient, random.UuidString, time.Now)

	donorStore := &donorStore{
		dynamoClient:  lpaDynamoClient,
		eventClient:   eventClient,
		logger:        logger,
		uuidString:    uuid.NewString,
		now:           time.Now,
		documentStore: documentStore,
	}
	certificateProviderStore := &certificateProviderStore{dynamoClient: lpaDynamoClient, now: time.Now}
	attorneyStore := &attorneyStore{dynamoClient: lpaDynamoClient, now: time.Now}
	shareCodeStore := &shareCodeStore{dynamoClient: lpaDynamoClient}
	dashboardStore := &dashboardStore{dynamoClient: lpaDynamoClient}
	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: lpaDynamoClient}

	shareCodeSender := page.NewShareCodeSender(shareCodeStore, notifyClient, appPublicURL, random.String)
	witnessCodeSender := page.NewWitnessCodeSender(donorStore, notifyClient)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger)

	rootMux := http.NewServeMux()
	handleRoot := makeHandle(rootMux, errorHandler, sessionStore)

	handleRoot(paths.Root, None,
		notFoundHandler)
	handleRoot(paths.SignOut, None,
		page.SignOut(logger, sessionStore, oneLoginClient, appPublicURL))
	handleRoot(paths.Fixtures, None,
		fixtures.Donor(tmpls.Get("fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, documentStore, eventClient))
	handleRoot(paths.CertificateProviderFixtures, None,
		fixtures.CertificateProvider(tmpls.Get("certificate_provider_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore))
	handleRoot(paths.AttorneyFixtures, None,
		fixtures.Attorney(tmpls.Get("attorney_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, attorneyStore))
	handleRoot(paths.DashboardFixtures, None,
		fixtures.Dashboard(tmpls.Get("dashboard_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, attorneyStore))
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
	handleRoot(page.Paths.LpaDeleted, RequireSession,
		page.Guidance(tmpls.Get("lpa_deleted.gohtml")))
	handleRoot(page.Paths.LpaWithdrawn, RequireSession,
		page.Guidance(tmpls.Get("lpa_withdrawn.gohtml")))

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		donorStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		certificateProviderStore,
		notFoundHandler,
		addressClient,
		notifyClient,
		shareCodeSender,
		dashboardStore,
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
		shareCodeStore,
		errorHandler,
		notFoundHandler,
		dashboardStore,
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
		shareCodeSender,
		witnessCodeSender,
		errorHandler,
		notFoundHandler,
		certificateProviderStore,
		attorneyStore,
		notifyClient,
		evidenceReceivedStore,
		documentStore,
		eventClient,
		dashboardStore,
		lpaStoreClient,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.String, errorHandler), localizer, lang, rumConfig, staticHash, oneloginURL)
}

func withAppData(next http.Handler, localizer page.Localizer, lang localize.Lang, rumConfig page.RumConfig, staticHash, oneloginURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";"); contentType != "multipart/form-data" {
			localizer.SetShowTranslationKeys(r.FormValue("showTranslationKeys") == "1")
		}

		appData := page.AppDataFromContext(ctx)
		appData.Path = r.URL.Path
		appData.Query = queryString(r)
		appData.Localizer = localizer
		appData.Lang = lang
		appData.RumConfig = rumConfig
		appData.StaticHash = staticHash
		appData.Paths = page.Paths
		appData.ActorTypes = actor.ActorTypes
		appData.OneloginURL = oneloginURL

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

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler, store sesh.Store) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()

			if opt&RequireSession != 0 {
				loginSession, err := sesh.Login(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
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
