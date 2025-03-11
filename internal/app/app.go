// Package app provides the web server for modernising-lpa.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneypage"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderpage"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/document"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donorpage"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/fixtures"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterpage"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherpage"
)

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type DynamoClient interface {
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK, v interface{}) error
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	AnyByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	BatchPut(ctx context.Context, items []interface{}) error
	Create(ctx context.Context, v interface{}) error
	CreateOnly(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v interface{}) error
	Move(ctx context.Context, oldKeys dynamo.Keys, value any) error
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	Put(ctx context.Context, v interface{}) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
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

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

func App(
	devMode bool,
	logger *slog.Logger,
	bundle Bundle,
	lang localize.Lang,
	tmpls, donorTmpls, certificateProviderTmpls, attorneyTmpls, supporterTmpls, voucherTmpls, guidanceTmpls template.Templates,
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
	useURL string,
) http.Handler {
	localizer := bundle.For(lang)
	documentStore := document.NewStore(lpaDynamoClient, s3Client, eventClient)

	donorStore := donor.NewStore(lpaDynamoClient, eventClient, logger, searchClient)
	certificateProviderStore := certificateprovider.NewStore(lpaDynamoClient)
	attorneyStore := attorney.NewStore(lpaDynamoClient)
	shareCodeStore := sharecode.NewStore(lpaDynamoClient)
	dashboardStore := dashboard.NewStore(lpaDynamoClient, lpastore.NewResolvingService(donorStore, lpaStoreClient))
	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: lpaDynamoClient}
	organisationStore := supporter.NewOrganisationStore(lpaDynamoClient)
	memberStore := supporter.NewMemberStore(lpaDynamoClient)
	voucherStore := voucher.NewStore(lpaDynamoClient)
	scheduledStore := scheduled.NewStore(lpaDynamoClient)
	progressTracker := task.ProgressTracker{Localizer: localizer}

	shareCodeSender := sharecode.NewSender(shareCodeStore, notifyClient, appPublicURL, eventClient, certificateProviderStore, scheduledStore)
	witnessCodeSender := donor.NewWitnessCodeSender(donorStore, certificateProviderStore, notifyClient, localizer)

	lpaStoreResolvingService := lpastore.NewResolvingService(donorStore, lpaStoreClient)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger, devMode)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger)

	rootMux := http.NewServeMux()
	handleRoot := makeHandle(rootMux, errorHandler, sessionStore)

	if devMode {
		handleRoot(page.PathFixtures, None,
			fixtures.Donor(tmpls.Get("fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient, shareCodeStore, voucherStore))
		handleRoot(page.PathCertificateProviderFixtures, None,
			fixtures.CertificateProvider(tmpls.Get("certificate_provider_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, eventClient, lpaStoreClient, lpaDynamoClient, organisationStore, memberStore, shareCodeStore))
		handleRoot(page.PathAttorneyFixtures, None,
			fixtures.Attorney(tmpls.Get("attorney_fixtures.gohtml"), sessionStore, shareCodeSender, donorStore, certificateProviderStore, attorneyStore, eventClient, lpaStoreClient, organisationStore, memberStore, shareCodeStore, lpaDynamoClient))
		handleRoot(page.PathSupporterFixtures, None,
			fixtures.Supporter(tmpls.Get("supporter_fixtures.gohtml"), sessionStore, organisationStore, donorStore, memberStore, lpaDynamoClient, searchClient, shareCodeStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient, voucherStore))
		handleRoot(page.PathVoucherFixtures, None,
			fixtures.Voucher(tmpls.Get("voucher_fixtures.gohtml"), sessionStore, shareCodeStore, shareCodeSender, donorStore, voucherStore, lpaStoreClient))
		handleRoot(page.PathDashboardFixtures, None,
			fixtures.Dashboard(tmpls.Get("dashboard_fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, shareCodeStore))
	}

	handleRoot(page.PathRoot, None,
		notFoundHandler)
	handleRoot(page.PathSignOut, None,
		page.SignOut(logger, sessionStore, oneLoginClient, appPublicURL))
	handleRoot(page.PathStart, None,
		page.Guidance(tmpls.Get("start.gohtml")))
	handleRoot(page.PathCertificateProviderStart, None,
		page.Guidance(tmpls.Get("certificate_provider_start.gohtml")))
	handleRoot(page.PathAttorneyStart, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml")))
	handleRoot(page.PathVoucherStart, None,
		page.Guidance(tmpls.Get("voucher_start.gohtml")))
	handleRoot(page.PathDashboard, RequireSession,
		page.Dashboard(tmpls.Get("dashboard.gohtml"), donorStore, dashboardStore, useURL))
	handleRoot(page.PathLpaDeleted, RequireSession,
		page.Guidance(tmpls.Get("lpa_deleted.gohtml")))
	handleRoot(page.PathLpaWithdrawn, RequireSession,
		page.Guidance(tmpls.Get("lpa_withdrawn.gohtml")))
	handleRoot(page.PathAccessibilityStatement, None,
		page.Guidance(tmpls.Get("accessibility_statement.gohtml")))

	handleRoot(page.PathAddingRestrictionsAndConditions, None,
		page.Guidance(guidanceTmpls.Get("adding_restrictions_and_conditions.gohtml")))
	handleRoot(page.PathContactTheOfficeOfThePublicGuardian, None,
		page.Guidance(guidanceTmpls.Get("contact_opg.gohtml")))
	handleRoot(page.PathHowDecisionsAreMadeWithMultipleAttorneys, None,
		page.Guidance(guidanceTmpls.Get("how_decisions_are_made_with_multiple_attorneys.gohtml")))
	handleRoot(page.PathHowToMakeAndRegisterYourLPA, None,
		page.Guidance(guidanceTmpls.Get("how_to_make_and_register_your_lpa.gohtml")))
	handleRoot(page.PathHowToSelectAttorneysForAnLPA, None,
		page.Guidance(guidanceTmpls.Get("how_to_select_attorneys_for_an_lpa.gohtml")))
	handleRoot(page.PathReplacementAttorneys, None,
		page.Guidance(guidanceTmpls.Get("replacement_attorneys.gohtml")))
	handleRoot(page.PathTheTwoTypesOfLPAPath, None,
		page.Guidance(guidanceTmpls.Get("the_two_types_of_lpa.gohtml")))
	handleRoot(page.PathUnderstandingLifeSustainingTreatment, None,
		page.Guidance(guidanceTmpls.Get("understanding_life_sustaining_treatment.gohtml")))
	handleRoot(page.PathUnderstandingMentalCapacity, None,
		page.Guidance(guidanceTmpls.Get("understanding_mental_capacity.gohtml")))

	voucherpage.Register(
		rootMux,
		logger,
		voucherTmpls,
		sessionStore,
		voucherStore,
		oneLoginClient,
		shareCodeStore,
		dashboardStore,
		errorHandler,
		lpaStoreResolvingService,
		notifyClient,
		appPublicURL,
		donorStore,
		lpaStoreClient,
		scheduledStore,
	)

	supporterpage.Register(
		rootMux,
		logger,
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
		progressTracker,
		lpaStoreResolvingService,
	)

	certificateproviderpage.Register(
		rootMux,
		logger,
		tmpls,
		certificateProviderTmpls,
		sessionStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		certificateProviderStore,
		addressClient,
		notifyClient,
		shareCodeSender,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
		donorStore,
		eventClient,
		scheduledStore,
		appPublicURL,
	)

	attorneypage.Register(
		rootMux,
		logger,
		tmpls,
		attorneyTmpls,
		sessionStore,
		attorneyStore,
		oneLoginClient,
		shareCodeStore,
		errorHandler,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
		notifyClient,
	)

	donorpage.Register(
		rootMux,
		logger,
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
		notifyClient,
		evidenceReceivedStore,
		documentStore,
		eventClient,
		dashboardStore,
		lpaStoreClient,
		shareCodeStore,
		progressTracker,
		lpaStoreResolvingService,
		scheduledStore,
		voucherStore,
		bundle,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.String, errorHandler), localizer, lang)
}

func withAppData(next http.Handler, localizer localize.Localizer, lang localize.Lang) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";"); contentType != "multipart/form-data" {
			localizer.SetShowTranslationKeys(r.FormValue("showTranslationKeys") == "1")
		}

		appData := appcontext.DataFromContext(ctx)
		appData.Path = r.URL.Path
		appData.Query = r.URL.Query()
		appData.Localizer = localizer
		appData.Lang = lang

		_, cookieErr := r.Cookie("cookies-consent")
		appData.CookieConsentSet = cookieErr != http.ErrNoCookie

		next.ServeHTTP(w, r.WithContext(appcontext.ContextWithData(ctx, appData)))
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

			appData := appcontext.DataFromContext(ctx)
			appData.Page = path.Format()

			if opt&RequireSession != 0 {
				loginSession, err := sessionStore.Login(r)
				if err != nil {
					http.Redirect(w, r, page.PathStart.Format(), http.StatusFound)
					return
				}

				appData.SessionID = loginSession.SessionID()
				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID})
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
