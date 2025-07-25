// Package app provides the web server for modernising-lpa.
package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/reuse"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
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
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	AllBySK(ctx context.Context, sk dynamo.SK, v any) error
	AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK dynamo.SK) ([]dynamo.Keys, error)
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	BatchPut(ctx context.Context, items []any) error
	Create(ctx context.Context, v any) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v any) error
	Move(ctx context.Context, oldKeys dynamo.Keys, value any) error
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	OneByPK(ctx context.Context, pk dynamo.PK, v any) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v any) error
	OneBySK(ctx context.Context, sk dynamo.SK, v any) error
	OneByUID(ctx context.Context, uid string) (dynamo.Keys, error)
	Put(ctx context.Context, v any) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, names map[string]string, values map[string]dynamodbtypes.AttributeValue, expression string) error
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
	donorStartURL string,
	certificateProviderStartURL string,
	attorneyStartURL string,
	useTestWitnessCode bool,
) http.Handler {
	localizer := bundle.For(lang)
	documentStore := document.NewStore(lpaDynamoClient, s3Client, eventClient)

	donorStore := donor.NewStore(lpaDynamoClient, eventClient, logger, searchClient)
	certificateProviderStore := certificateprovider.NewStore(lpaDynamoClient)
	attorneyStore := attorney.NewStore(lpaDynamoClient)
	accessCodeStore := accesscode.NewStore(lpaDynamoClient)
	dashboardStore := dashboard.NewStore(lpaDynamoClient, lpastore.NewResolvingService(donorStore, lpaStoreClient))
	evidenceReceivedStore := &evidenceReceivedStore{dynamoClient: lpaDynamoClient}
	organisationStore := supporter.NewOrganisationStore(lpaDynamoClient)
	memberStore := supporter.NewMemberStore(lpaDynamoClient)
	voucherStore := voucher.NewStore(lpaDynamoClient)
	scheduledStore := scheduled.NewStore(lpaDynamoClient)
	reuseStore := reuse.NewStore(lpaDynamoClient)
	progressTracker := task.ProgressTracker{Localizer: localizer}

	accessCodeSender := accesscode.NewSender(accessCodeStore, notifyClient, appPublicURL, certificateProviderStartURL, attorneyStartURL, eventClient, certificateProviderStore, scheduledStore)
	witnessCodeSender := donor.NewWitnessCodeSender(donorStore, certificateProviderStore, notifyClient, localizer, useTestWitnessCode)

	lpaStoreResolvingService := lpastore.NewResolvingService(donorStore, lpaStoreClient)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger, devMode)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger, donorStartURL)

	rootMux := http.NewServeMux()
	handleRoot := makeHandle(rootMux, errorHandler, sessionStore, donorStartURL)

	if devMode {
		handleRoot(page.PathFixtures, None,
			fixtures.Donor(tmpls.Get("fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient, accessCodeStore, voucherStore, reuseStore, notifyClient, appPublicURL))
		handleRoot(page.PathCertificateProviderFixtures, None,
			fixtures.CertificateProvider(tmpls.Get("certificate_provider_fixtures.gohtml"), sessionStore, accessCodeSender, donorStore, certificateProviderStore, eventClient, lpaStoreClient, lpaDynamoClient, organisationStore, memberStore, accessCodeStore))
		handleRoot(page.PathAttorneyFixtures, None,
			fixtures.Attorney(tmpls.Get("attorney_fixtures.gohtml"), sessionStore, accessCodeSender, donorStore, certificateProviderStore, attorneyStore, eventClient, lpaStoreClient, organisationStore, memberStore, accessCodeStore, lpaDynamoClient))
		handleRoot(page.PathSupporterFixtures, None,
			fixtures.Supporter(tmpls.Get("supporter_fixtures.gohtml"), sessionStore, organisationStore, donorStore, memberStore, lpaDynamoClient, searchClient, accessCodeStore, certificateProviderStore, attorneyStore, documentStore, eventClient, lpaStoreClient, voucherStore, reuseStore, notifyClient, appPublicURL))
		handleRoot(page.PathVoucherFixtures, None,
			fixtures.Voucher(tmpls.Get("voucher_fixtures.gohtml"), sessionStore, accessCodeStore, accessCodeSender, donorStore, voucherStore, lpaStoreClient))
		handleRoot(page.PathDashboardFixtures, None,
			fixtures.Dashboard(tmpls.Get("dashboard_fixtures.gohtml"), sessionStore, donorStore, certificateProviderStore, attorneyStore, accessCodeStore))
	}

	handleRoot(page.PathRoot, None,
		notFoundHandler)
	handleRoot(page.PathSignOut, None,
		page.SignOut(logger, sessionStore, oneLoginClient, donorStartURL))
	handleRoot(page.PathStart, None,
		page.Guidance(tmpls.Get("start.gohtml")))
	handleRoot(page.PathCertificateProviderStart, None,
		page.Guidance(tmpls.Get("certificate_provider_start.gohtml")))
	handleRoot(page.PathAttorneyStart, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml")))
	handleRoot(page.PathVoucherStart, None,
		page.Guidance(tmpls.Get("voucher_start.gohtml")))
	handleRoot(page.PathDashboard, RequireSession,
		page.Dashboard(tmpls.Get("dashboard.gohtml"), dashboardStore, useURL, sessionStore))
	handleRoot(page.PathMakeOrAddAnLPA, RequireSession,
		page.MakeOrAddAnLPA(tmpls.Get("make_or_add_an_lpa.gohtml"), donorStore, dashboardStore, eventClient))
	handleRoot(page.PathAddAnLPA, RequireSession,
		page.AddAnLPA(tmpls.Get("add_an_lpa.gohtml")))
	handleRoot(page.PathLpaDeleted, RequireSession,
		page.Guidance(tmpls.Get("lpa_deleted.gohtml")))
	handleRoot(page.PathLpaWithdrawn, RequireSession,
		page.Guidance(tmpls.Get("lpa_withdrawn.gohtml")))
	handleRoot(page.PathAccessibilityStatement, None,
		page.Guidance(tmpls.Get("accessibility_statement.gohtml")))

	handleRoot(page.PathAddingRestrictionsAndConditions, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("adding_restrictions_and_conditions.gohtml")))
	handleRoot(page.PathContactTheOfficeOfThePublicGuardian, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("contact_opg.gohtml")))
	handleRoot(page.PathGlossary, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("glossary.gohtml")))
	handleRoot(page.PathHowDecisionsAreMadeWithMultipleAttorneys, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("how_decisions_are_made_with_multiple_attorneys.gohtml")))
	handleRoot(page.PathHowToMakeAndRegisterYourLPA, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("how_to_make_and_register_your_lpa.gohtml")))
	handleRoot(page.PathHowToSelectAttorneysForAnLPA, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("how_to_select_attorneys_for_an_lpa.gohtml")))
	handleRoot(page.PathReplacementAttorneys, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("replacement_attorneys.gohtml")))
	handleRoot(page.PathTheTwoTypesOfLPAPath, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("the_two_types_of_lpa.gohtml")))
	handleRoot(page.PathUnderstandingLifeSustainingTreatment, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("understanding_life_sustaining_treatment.gohtml")))
	handleRoot(page.PathUnderstandingMentalCapacity, RequireSession|HideNav,
		page.Guidance(guidanceTmpls.Get("understanding_mental_capacity.gohtml")))

	voucherpage.Register(
		rootMux,
		logger,
		voucherTmpls,
		sessionStore,
		voucherStore,
		oneLoginClient,
		accessCodeStore,
		dashboardStore,
		errorHandler,
		lpaStoreResolvingService,
		notifyClient,
		donorStore,
		lpaStoreClient,
		scheduledStore,
		donorStartURL,
		localizer,
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
		accessCodeStore,
		progressTracker,
		lpaStoreResolvingService,
		donorStartURL,
	)

	certificateproviderpage.Register(
		rootMux,
		logger,
		tmpls,
		certificateProviderTmpls,
		sessionStore,
		oneLoginClient,
		accessCodeStore,
		errorHandler,
		certificateProviderStore,
		addressClient,
		notifyClient,
		accessCodeSender,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
		donorStore,
		eventClient,
		scheduledStore,
		bundle,
		donorStartURL,
		certificateProviderStartURL,
	)

	attorneypage.Register(
		rootMux,
		logger,
		tmpls,
		attorneyTmpls,
		sessionStore,
		attorneyStore,
		oneLoginClient,
		accessCodeStore,
		errorHandler,
		dashboardStore,
		lpaStoreClient,
		lpaStoreResolvingService,
		notifyClient,
		eventClient,
		bundle,
		attorneyStartURL,
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
		accessCodeSender,
		witnessCodeSender,
		errorHandler,
		certificateProviderStore,
		notifyClient,
		evidenceReceivedStore,
		documentStore,
		eventClient,
		dashboardStore,
		lpaStoreClient,
		accessCodeStore,
		progressTracker,
		lpaStoreResolvingService,
		scheduledStore,
		voucherStore,
		reuseStore,
		bundle,
		donorStartURL,
		certificateProviderStartURL,
		attorneyStartURL,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.AlphaNumeric, errorHandler), localizer, lang)
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
	HideNav
)

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler, sessionStore SessionStore, donorStartURL string) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.Page = path.Format()
			appData.HideLoginNav = opt&HideNav != 0

			if opt&RequireSession != 0 {
				loginSession, err := sessionStore.Login(r)
				if err != nil {
					http.Redirect(w, r, donorStartURL, http.StatusFound)
					return
				}

				appData.SessionID = loginSession.SessionID()

				if !appData.HasLpas {
					appData.HasLpas = loginSession.HasLPAs
				}

				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID})
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
