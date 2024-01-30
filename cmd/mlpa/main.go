package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	html "html/template"
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
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/templatefn"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/mod/sumdb/dirhash"
)

var Tag string

func main() {
	ctx := context.Background()
	logger := logging.New(os.Stdout, "opg-modernising-lpa2")

	var (
		appPublicURL          = env.Get("APP_PUBLIC_URL", "http://localhost:5050")
		authRedirectBaseURL   = env.Get("AUTH_REDIRECT_BASE_URL", "http://localhost:5050")
		webDir                = env.Get("WEB_DIR", "web")
		awsBaseURL            = env.Get("AWS_BASE_URL", "")
		clientID              = env.Get("CLIENT_ID", "client-id-value")
		issuer                = env.Get("ISSUER", "http://mock-onelogin:8080")
		dynamoTableLpas       = env.Get("DYNAMODB_TABLE_LPAS", "lpas")
		notifyBaseURL         = env.Get("GOVUK_NOTIFY_BASE_URL", "http://mock-notify:8080")
		notifyIsProduction    = env.Get("GOVUK_NOTIFY_IS_PRODUCTION", "") == "1"
		ordnanceSurveyBaseURL = env.Get("ORDNANCE_SURVEY_BASE_URL", "http://mock-os-api:8080")
		payBaseURL            = env.Get("GOVUK_PAY_BASE_URL", "http://mock-pay:8080")
		port                  = env.Get("APP_PORT", "8080")
		xrayEnabled           = env.Get("XRAY_ENABLED", "") == "1"
		rumConfig             = page.RumConfig{
			GuestRoleArn:      env.Get("AWS_RUM_GUEST_ROLE_ARN", ""),
			Endpoint:          env.Get("AWS_RUM_ENDPOINT", ""),
			ApplicationRegion: env.Get("AWS_RUM_APPLICATION_REGION", ""),
			IdentityPoolID:    env.Get("AWS_RUM_IDENTITY_POOL_ID", ""),
			ApplicationID:     env.Get("AWS_RUM_APPLICATION_ID", ""),
		}
		uidBaseURL            = env.Get("UID_BASE_URL", "http://mock-uid:8080")
		lpaStoreBaseURL       = env.Get("LPA_STORE_BASE_URL", "http://mock-lpa-store:8080")
		metadataURL           = env.Get("ECS_CONTAINER_METADATA_URI_V4", "")
		oneloginURL           = env.Get("ONELOGIN_URL", "https://home.integration.account.gov.uk")
		evidenceBucketName    = env.Get("UPLOADS_S3_BUCKET_NAME", "evidence")
		eventBusName          = env.Get("EVENT_BUS_NAME", "default")
		mockIdentityPublicKey = env.Get("MOCK_IDENTITY_PUBLIC_KEY", "")
	)

	staticHash, err := dirhash.HashDir(webDir+"/static", webDir, dirhash.DefaultHash)
	if err != nil {
		logger.Fatal(err)
	}
	staticHash = url.QueryEscape(staticHash[3:11])

	httpClient := &http.Client{Timeout: 10 * time.Second}

	if xrayEnabled {
		resource, err := ecs.NewResourceDetector().Detect(ctx)
		if err != nil {
			logger.Fatal(err)
		}

		shutdown, err := telemetry.Setup(ctx, resource)
		if err != nil {
			logger.Fatal(err)
		}
		defer shutdown(ctx)

		httpClient.Transport = otelhttp.NewTransport(httpClient.Transport)
	}

	var region string
	if metadataURL != "" {
		region, err = awsRegion(metadataURL)
		if err != nil {
			logger.Print("error getting region:", err)
		}
	}

	layouts, err := parseLayoutTemplates(webDir+"/template/layout", templatefn.All(Tag, region))
	if err != nil {
		logger.Fatal(err)
	}

	tmpls, err := parseTemplates(webDir+"/template", layouts)
	if err != nil {
		logger.Fatal(err)
	}

	donorTmpls, err := parseTemplates(webDir+"/template/donor", layouts)
	if err != nil {
		logger.Fatal(err)
	}

	certificateProviderTmpls, err := parseTemplates(webDir+"/template/certificateprovider", layouts)
	if err != nil {
		logger.Fatal(err)
	}

	attorneyTmpls, err := parseTemplates(webDir+"/template/attorney", layouts)
	if err != nil {
		logger.Fatal(err)
	}

	supporterTmpls, err := parseTemplates(webDir+"/template/supporter", layouts)
	if err != nil {
		logger.Fatal(err)
	}

	bundle, err := localize.NewBundle("lang/en.json", "lang/cy.json")
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal(fmt.Errorf("unable to load SDK config: %w", err))
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	if len(awsBaseURL) > 0 {
		cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsBaseURL,
				SigningRegion: "eu-west-1",
			}, nil
		})
	}

	lpasDynamoClient, err := dynamo.NewClient(cfg, dynamoTableLpas)
	if err != nil {
		logger.Fatal(err)
	}

	eventClient := event.NewClient(cfg, eventBusName)

	secretsClient, err := secrets.NewClient(cfg, time.Hour)
	if err != nil {
		logger.Fatal(err)
	}

	sessionKeys, err := secretsClient.CookieSessionKeys(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	sessionStore := sessions.NewCookieStore(sessionKeys...)

	redirectURL := authRedirectBaseURL + page.Paths.AuthRedirect.Format()

	identityPublicKeyFunc := func(ctx context.Context) (*ecdsa.PublicKey, error) {
		bytes, err := secretsClient.SecretBytes(ctx, secrets.GovUkOneLoginIdentityPublicKey)
		if err != nil {
			return nil, err
		}

		return jwt.ParseECPublicKeyFromPEM(bytes)
	}

	if mockIdentityPublicKey != "" {
		identityPublicKeyFunc = func(ctx context.Context) (*ecdsa.PublicKey, error) {
			bytes, err := base64.StdEncoding.DecodeString(mockIdentityPublicKey)
			if err != nil {
				return nil, err
			}

			return jwt.ParseECPublicKeyFromPEM(bytes)
		}
	}

	oneloginClient := onelogin.New(ctx, logger, httpClient, secretsClient, issuer, clientID, redirectURL, identityPublicKeyFunc)

	payApiKey, err := secretsClient.Secret(ctx, secrets.GovUkPay)
	if err != nil {
		logger.Fatal(err)
	}

	payClient := &pay.Client{
		BaseURL:    payBaseURL,
		ApiKey:     payApiKey,
		HttpClient: httpClient,
	}

	osApiKey, err := secretsClient.Secret(ctx, secrets.OrdnanceSurvey)
	if err != nil {
		logger.Fatal(err)
	}

	addressClient := place.NewClient(ordnanceSurveyBaseURL, osApiKey, httpClient)

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		logger.Fatal(err)
	}

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, httpClient)
	if err != nil {
		logger.Fatal(err)
	}

	evidenceS3Client := s3.NewClient(cfg, evidenceBucketName)

	lambdaClient := lambda.New(cfg, v4.NewSigner(), httpClient, time.Now)
	uidClient := uid.New(uidBaseURL, lambdaClient)
	lpaStoreClient := lpastore.New(lpaStoreBaseURL, secretsClient, lambdaClient)

	mux := http.NewServeMux()
	mux.HandleFunc(page.Paths.HealthCheck.Service.String(), func(w http.ResponseWriter, r *http.Request) {})
	mux.Handle(page.Paths.HealthCheck.Dependency.String(), page.DependencyHealthCheck(logger, map[string]page.HealthChecker{
		"uid":      uidClient,
		"onelogin": oneloginClient,
		"lpaStore": lpaStoreClient,
	}))
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"/robots.txt")
	})
	mux.Handle("/static/", http.StripPrefix("/static", handlers.CompressHandler(page.CacheControlHeaders(http.FileServer(http.Dir(webDir+"/static/"))))))
	mux.Handle(page.Paths.AuthRedirect.String(), page.AuthRedirect(logger, sessionStore))
	mux.Handle(page.Paths.CookiesConsent.String(), page.CookieConsent(page.Paths))

	mux.Handle("/cy/", http.StripPrefix("/cy", app.App(
		logger,
		bundle.For(localize.Cy),
		localize.Cy,
		tmpls,
		donorTmpls,
		certificateProviderTmpls,
		attorneyTmpls,
		supporterTmpls,
		sessionStore,
		lpasDynamoClient,
		appPublicURL,
		payClient,
		notifyClient,
		addressClient,
		rumConfig,
		staticHash,
		page.Paths,
		oneloginClient,
		oneloginURL,
		evidenceS3Client,
		eventClient,
		lpaStoreClient,
	)))

	mux.Handle("/", app.App(
		logger,
		bundle.For(localize.En),
		localize.En,
		tmpls,
		donorTmpls,
		certificateProviderTmpls,
		attorneyTmpls,
		supporterTmpls,
		sessionStore,
		lpasDynamoClient,
		appPublicURL,
		payClient,
		notifyClient,
		addressClient,
		rumConfig,
		staticHash,
		page.Paths,
		oneloginClient,
		oneloginURL,
		evidenceS3Client,
		eventClient,
		lpaStoreClient,
	))

	var handler http.Handler = mux
	if xrayEnabled {
		handler = telemetry.WrapHandler(mux)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           page.Recover(tmpls.Get("error-500.gohtml"), logger, bundle, handler),
		ReadHeaderTimeout: 20 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	logger.Print("Running at :" + port)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Print("signal received: ", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		logger.Print(err)
	}
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
