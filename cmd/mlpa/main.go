package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/handlers"
	"github.com/ministryofjustice/opg-go-common/securityheaders"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/app"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lambda"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/templatefn"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/mod/sumdb/dirhash"
)

var Tag string

func main() {
	ctx := context.Background()

	handler := telemetry.NewSlogHandler(slog.
		NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
				switch a.Value.Kind() {
				case slog.KindAny:
					switch v := a.Value.Any().(type) {
					case *http.Request:
						return slog.Group(a.Key,
							slog.String("method", v.Method),
							slog.String("uri", v.URL.String()))
					}
				}

				return a
			},
		}))

	logger := slog.New(handler.
		WithAttrs([]slog.Attr{
			slog.String("service_name", "opg-modernising-lpa"),
			slog.String("tag", Tag),
		}))

	if err := run(ctx, logger); err != nil {
		logger.Error("run error", slog.Any("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	var (
		devMode               = os.Getenv("DEV_MODE") == "1"
		appPublicURL          = cmp.Or(os.Getenv("APP_PUBLIC_URL"), "http://localhost:5050")
		authRedirectBaseURL   = cmp.Or(os.Getenv("AUTH_REDIRECT_BASE_URL"), "http://localhost:5050")
		webDir                = cmp.Or(os.Getenv("WEB_DIR"), "web")
		awsBaseURL            = os.Getenv("AWS_BASE_URL")
		clientID              = cmp.Or(os.Getenv("CLIENT_ID"), "client-id-value")
		issuer                = cmp.Or(os.Getenv("ISSUER"), "http://mock-onelogin:8080")
		identityURL           = cmp.Or(os.Getenv("IDENTITY_URL"), "http://mock-onelogin:8080")
		dynamoTableLpas       = cmp.Or(os.Getenv("DYNAMODB_TABLE_LPAS"), "lpas")
		notifyBaseURL         = cmp.Or(os.Getenv("GOVUK_NOTIFY_BASE_URL"), "http://mock-notify:8080")
		notifyIsProduction    = os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
		ordnanceSurveyBaseURL = cmp.Or(os.Getenv("ORDNANCE_SURVEY_BASE_URL"), "http://mock-os-api:8080")
		payBaseURL            = cmp.Or(os.Getenv("GOVUK_PAY_BASE_URL"), "http://mock-pay:8080")
		port                  = cmp.Or(os.Getenv("APP_PORT"), "8080")
		xrayEnabled           = os.Getenv("XRAY_ENABLED") == "1"
		rumConfig             = templatefn.RumConfig{
			GuestRoleArn:      os.Getenv("AWS_RUM_GUEST_ROLE_ARN"),
			Endpoint:          os.Getenv("AWS_RUM_ENDPOINT"),
			ApplicationRegion: os.Getenv("AWS_RUM_APPLICATION_REGION"),
			IdentityPoolID:    os.Getenv("AWS_RUM_IDENTITY_POOL_ID"),
			ApplicationID:     os.Getenv("AWS_RUM_APPLICATION_ID"),
		}
		uidBaseURL            = cmp.Or(os.Getenv("UID_BASE_URL"), "http://mock-uid:8080")
		lpaStoreBaseURL       = cmp.Or(os.Getenv("LPA_STORE_BASE_URL"), "http://mock-lpa-store:8080")
		lpaStoreSecretARN     = os.Getenv("LPA_STORE_SECRET_ARN")
		metadataURL           = os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
		oneloginURL           = cmp.Or(os.Getenv("ONELOGIN_URL"), "https://home.integration.account.gov.uk")
		evidenceBucketName    = cmp.Or(os.Getenv("UPLOADS_S3_BUCKET_NAME"), "evidence")
		eventBusName          = cmp.Or(os.Getenv("EVENT_BUS_NAME"), "default")
		searchEndpoint        = os.Getenv("SEARCH_ENDPOINT")
		searchIndexName       = cmp.Or(os.Getenv("SEARCH_INDEX_NAME"), "lpas")
		searchIndexingEnabled = os.Getenv("SEARCH_INDEXING_DISABLED") != "1"
	)

	staticHash, err := dirhash.HashDir(webDir+"/static", webDir, dirhash.DefaultHash)
	if err != nil {
		return err
	}
	staticHash = url.QueryEscape(staticHash[3:11])

	httpClient := &http.Client{Timeout: 30 * time.Second}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	if len(awsBaseURL) > 0 {
		cfg.BaseEndpoint = aws.String(awsBaseURL)
	}

	if xrayEnabled {
		resource, err := ecs.NewResourceDetector().Detect(ctx)
		if err != nil {
			return err
		}

		shutdown, err := telemetry.Setup(ctx, resource, &cfg.APIOptions)
		if err != nil {
			return err
		}
		defer shutdown(ctx)

		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)
	}

	var region string
	if metadataURL != "" {
		region, err = awsRegion(metadataURL)
		if err != nil {
			logger.Warn("error getting region:", slog.Any("err", err))
		}
	}

	layouts, err := parseLayoutTemplates(webDir+"/template/layout", templatefn.All(&templatefn.Globals{
		DevMode:     devMode,
		Tag:         Tag,
		Region:      region,
		OneloginURL: oneloginURL,
		StaticHash:  staticHash,
		RumConfig:   rumConfig,
		ActorTypes:  actor.TypeValues,
	}))
	if err != nil {
		return err
	}

	tmpls, err := parseTemplates(webDir+"/template", layouts)
	if err != nil {
		return err
	}

	donorTmpls, err := parseTemplates(webDir+"/template/donor", layouts)
	if err != nil {
		return err
	}

	certificateProviderTmpls, err := parseTemplates(webDir+"/template/certificateprovider", layouts)
	if err != nil {
		return err
	}

	attorneyTmpls, err := parseTemplates(webDir+"/template/attorney", layouts)
	if err != nil {
		return err
	}

	supporterTmpls, err := parseTemplates(webDir+"/template/supporter", layouts)
	if err != nil {
		return err
	}

	voucherTmpls, err := parseTemplates(webDir+"/template/voucher", layouts)
	if err != nil {
		return err
	}

	bundle, err := localize.NewBundle("lang/en.json", "lang/cy.json")
	if err != nil {
		return err
	}

	lpasDynamoClient, err := dynamo.NewClient(cfg, dynamoTableLpas)
	if err != nil {
		return err
	}

	eventClient := event.NewClient(cfg, eventBusName)

	searchClient, err := search.NewClient(cfg, searchEndpoint, searchIndexName, searchIndexingEnabled)
	if err != nil {
		return err
	}

	if err := searchClient.CreateIndices(ctx); err != nil {
		return err
	}

	secretsClient, err := secrets.NewClient(cfg, time.Hour)
	if err != nil {
		return err
	}

	sessionKeys, err := secretsClient.CookieSessionKeys(ctx)
	if err != nil {
		return err
	}

	sessionStore := sesh.NewStore(lpasDynamoClient, sessionKeys)

	redirectURL := authRedirectBaseURL + page.PathAuthRedirect.Format()

	oneloginClient := onelogin.New(ctx, logger, httpClient, secretsClient, issuer, identityURL, clientID, redirectURL)

	payApiKey, err := secretsClient.Secret(ctx, secrets.GovUkPay)
	if err != nil {
		return err
	}

	payClient := pay.New(logger, httpClient, payBaseURL, payApiKey)

	osApiKey, err := secretsClient.Secret(ctx, secrets.OrdnanceSurvey)
	if err != nil {
		return err
	}

	addressClient := place.NewClient(ordnanceSurveyBaseURL, osApiKey, httpClient)

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		return err
	}

	notifyClient, err := notify.New(logger, notifyIsProduction, notifyBaseURL, notifyApiKey, httpClient, eventClient, bundle)
	if err != nil {
		return err
	}

	evidenceS3Client := s3.NewClient(cfg, evidenceBucketName)

	lambdaClient := lambda.New(cfg, v4.NewSigner(), httpClient, time.Now)
	uidClient := uid.New(uidBaseURL, lambdaClient)
	lpaStoreClient := lpastore.New(lpaStoreBaseURL, secretsClient, lpaStoreSecretARN, lambdaClient)

	mux := http.NewServeMux()
	mux.HandleFunc(page.PathHealthCheckService.String(), func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle(page.PathHealthCheckDependency.String(), page.DependencyHealthCheck(map[string]page.HealthChecker{
		"uid":      uidClient,
		"onelogin": oneloginClient,
		"lpaStore": lpaStoreClient,
	}))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"/robots.txt")
	})
	mux.Handle("/static/", http.StripPrefix("/static", handlers.CompressHandler(page.CacheControlHeaders(http.FileServer(http.Dir(webDir+"/static/"))))))
	mux.Handle(page.PathAuthRedirect.String(), page.AuthRedirect(logger, sessionStore))
	mux.Handle(page.PathCookiesConsent.String(), page.CookieConsent())

	mux.Handle("/cy/", http.StripPrefix("/cy", app.App(
		devMode,
		logger,
		bundle.For(localize.Cy),
		localize.Cy,
		tmpls,
		donorTmpls,
		certificateProviderTmpls,
		attorneyTmpls,
		supporterTmpls,
		voucherTmpls,
		sessionStore,
		lpasDynamoClient,
		appPublicURL,
		payClient,
		notifyClient,
		addressClient,
		oneloginClient,
		evidenceS3Client,
		eventClient,
		lpaStoreClient,
		searchClient,
	)))

	mux.Handle("/", app.App(
		devMode,
		logger,
		bundle.For(localize.En),
		localize.En,
		tmpls,
		donorTmpls,
		certificateProviderTmpls,
		attorneyTmpls,
		supporterTmpls,
		voucherTmpls,
		sessionStore,
		lpasDynamoClient,
		appPublicURL,
		payClient,
		notifyClient,
		addressClient,
		oneloginClient,
		evidenceS3Client,
		eventClient,
		lpaStoreClient,
		searchClient,
	))

	var handler http.Handler = mux
	if xrayEnabled {
		handler = telemetry.WrapHandler(mux)
	}
	handler = securityheaders.Use(handler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           page.Recover(tmpls.Get("error-500.gohtml"), logger, bundle, handler),
		ReadHeaderTimeout: 20 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error("listen and serve error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	logger.Info("started", slog.String("port", port))

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Info("signal received", slog.String("signal", sig.String()))

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return server.Shutdown(tc)
}

func awsRegion(metadataURL string) (string, error) {
	resp, err := http.Get(metadataURL + "/task")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var metadata struct{ TaskARN string }
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", err
	}

	parts := strings.Split(metadata.TaskARN, ":")
	if len(parts) < 4 {
		return "", fmt.Errorf("TaskARN contained only %d parts", len(parts))
	}

	return parts[3], nil
}

func parseLayoutTemplates(layoutDir string, funcs html.FuncMap) (*html.Template, error) {
	return html.New("").Funcs(funcs).ParseGlob(filepath.Join(layoutDir, "*.*"))
}

func parseTemplates(templateDir string, layouts *html.Template) (template.Templates, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.*"))
	if err != nil {
		return nil, err
	}

	tmpls := map[string]*html.Template{}
	for _, file := range files {
		clone, err := layouts.Clone()
		if err != nil {
			return nil, err
		}

		tmpl, err := clone.ParseFiles(file)
		if err != nil {
			return nil, err
		}

		tmpls[filepath.Base(file)] = tmpl
	}

	return tmpls, nil
}
