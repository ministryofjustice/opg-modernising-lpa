package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type DynamoClient interface {
	One(ctx context.Context, pk, sk string, v interface{}) error
	OneByPK(ctx context.Context, pk string, v interface{}) error
	OneByPartialSK(ctx context.Context, pk, partialSK string, v interface{}) error
	AllByPartialSK(ctx context.Context, pk, partialSK string, v interface{}) error
	LatestForActor(ctx context.Context, sk string, v interface{}) error
	AllBySK(ctx context.Context, sk string, v interface{}) error
	AllByKeys(ctx context.Context, keys []dynamo.Key) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllKeysByPK(ctx context.Context, pk string) ([]dynamo.Key, error)
	Put(ctx context.Context, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Key) error
	DeleteOne(ctx context.Context, pk, sk string) error
	Update(ctx context.Context, pk, sk string, values map[string]dynamodbtypes.AttributeValue, expression string) error
	BatchPut(ctx context.Context, items []interface{}) error
	OneBySK(ctx context.Context, sk string, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
}

type S3Client interface {
	PutObject(context.Context, string, []byte) error
	DeleteObject(context.Context, string) error
	DeleteObjects(ctx context.Context, keys []string) error
	PutObjectTagging(context.Context, string, map[string]string) error
}

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
}

func App(
	logger *slog.Logger,
	localizer page.Localizer,
	lang localize.Lang,
	tmpls, donorTmpls, certificateProviderTmpls, attorneyTmpls, supporterTmpls template.Templates,
	sessionStore *sesh.Store,
	lpaDynamoClient DynamoClient,
	appPublicURL string,
	payClient *pay.Client,
	notifyClient *notify.Client,
	addressClient *place.Client,
	oneLoginClient *onelogin.Client,
	s3Client S3Client,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
	searchClient *search.Client,
) http.Handler {
	documentStore := NewDocumentStore(lpaDynamoClient, s3Client, eventClient, random.UuidString, time.Now)

	donorStore := &donorStore{
		dynamoClient:  lpaDynamoClient,
		eventClient:   eventClient,
		logger:        logger,
		uuidString:    uuid.NewString,
		newUID:        actoruid.New,
		now:           time.Now,
		documentStore: documentStore,
		searchClient:  searchClient,
	}
	certificateProviderStore := &certificateProviderStore{dynamoClient: lpaDynamoClient, now: time.Now}
	attorneyStore := &attorneyStore{dynamoClient: lpaDynamoClient, now: time.Now}
	shareCodeStore := &shareCodeStore{dynamoClient: lpaDynamoClient, now: time.Now}
	dashboardStore := &dashboardStore{dynamoClient: lpaDynamoClient}
	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: lpaDynamoClient}
	organisationStore := &organisationStore{dynamoClient: lpaDynamoClient, now: time.Now, uuidString: uuid.NewString, newUID: actoruid.New}
	memberStore := &memberStore{dynamoClient: lpaDynamoClient, now: time.Now, uuidString: uuid.NewString}
	progressTracker := page.ProgressTracker{Localizer: localizer}

	shareCodeSender := page.NewShareCodeSender(shareCodeStore, notifyClient, appPublicURL, random.String, eventClient)
	witnessCodeSender := page.NewWitnessCodeSender(donorStore, notifyClient)

	lpaStoreResolvingService := lpastore.NewResolvingService(donorStore, lpaStoreClient)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger)

	rootMux := http.NewServeMux()
	handleRoot := makeHandle(rootMux, errorHandler, sessionStore)

	handleRoot(page.Paths.Root, None,
		notFoundHandler)
	handleRoot(page.Paths.SignOut, None,
		page.SignOut(logger, sessionStore, oneLoginClient, appPublicURL))
	handleRoot(page.Paths.Fixtures, None,
		fixtures.Donor(tmpls.Get("fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient))
	handleRoot(page.Paths.CertificateProviderFixtures, None,
		fixtures.CertificateProvider(tmpls.Get("certificate_provider_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, eventClient, lpaStoreClient, lpaDynamoClient))
	handleRoot(page.Paths.AttorneyFixtures, None,
		fixtures.Attorney(tmpls.Get("attorney_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, attorneyStore, eventClient, lpaStoreClient))
	handleRoot(page.Paths.SupporterFixtures, None,
		fixtures.Supporter(sessionStore, organisationStore, donorStore, memberStore, lpaDynamoClient, searchClient, shareCodeStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient))
	handleRoot(page.Paths.DashboardFixtures, None,
		fixtures.Dashboard(tmpls.Get("dashboard_fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore))
	handleRoot(page.Paths.YourLegalRightsAndResponsibilities, None,
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

	supporter.Register(
		rootMux,
		supporterTmpls,
		oneLoginClient,
		sessionStore,
		organisationStore,
		errorHandler,
		notifyClient,
		appPublicURL,
		memberStore,
		searchClient,
		donorStore,
		shareCodeStore,
		certificateProviderStore,
		attorneyStore,
		progressTracker,
		lpaStoreResolvingService,
	)

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		certificateProviderTmpls,
		sessionStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		certificateProviderStore,
		notFoundHandler,
		addressClient,
		notifyClient,
		shareCodeSender,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
	)

	attorney.Register(
		rootMux,
		logger,
		tmpls,
		attorneyTmpls,
		sessionStore,
		certificateProviderStore,
		attorneyStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		notFoundHandler,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
	)

	donor.Register(
		rootMux,
		logger,
		tmpls,
		donorTmpls,
		sessionStore,
		donorStore,
		oneLoginClient,
		addressClient,
		appPublicURL,
		payClient,
		shareCodeSender,
		witnessCodeSender,
		errorHandler,
		certificateProviderStore,
		attorneyStore,
		notifyClient,
		evidenceReceivedStore,
		documentStore,
		eventClient,
		dashboardStore,
		lpaStoreClient,
		shareCodeStore,
		progressTracker,
		lpaStoreResolvingService,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.String, errorHandler), localizer, lang)
}

func withAppData(next http.Handler, localizer page.Localizer, lang localize.Lang) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";"); contentType != "multipart/form-data" {
			localizer.SetShowTranslationKeys(r.FormValue("showTranslationKeys") == "1")
		}

		appData := page.AppDataFromContext(ctx)
		appData.Path = r.URL.Path
		appData.Query = r.URL.Query()
		appData.Localizer = localizer
		appData.Lang = lang
		appData.CanToggleWelsh = true

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

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler, sessionStore SessionStore) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()

			if opt&RequireSession != 0 {
				loginSession, err := sessionStore.Login(r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = loginSession.SessionID()
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
