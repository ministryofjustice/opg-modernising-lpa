package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/telemetry"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/templatefn"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/mod/sumdb/dirhash"
)

func main() {
	ctx := context.Background()
	logger := logging.New(os.Stdout, "opg-modernising-lpa")

	var (
		appPublicURL          = env.Get("APP_PUBLIC_URL", "http://localhost:5050")
		authRedirectBaseURL   = env.Get("AUTH_REDIRECT_BASE_URL", "http://localhost:5050")
		webDir                = env.Get("WEB_DIR", "web")
		awsBaseURL            = env.Get("AWS_BASE_URL", "")
		clientID              = env.Get("CLIENT_ID", "client-id-value")
		issuer                = env.Get("ISSUER", "http://sign-in-mock:7012")
		dynamoTableLpas       = env.Get("DYNAMODB_TABLE_LPAS", "")
		notifyBaseURL         = env.Get("GOVUK_NOTIFY_BASE_URL", "")
		notifyIsProduction    = env.Get("GOVUK_NOTIFY_IS_PRODUCTION", "") == "1"
		ordnanceSurveyBaseUrl = env.Get("ORDNANCE_SURVEY_BASE_URL", "http://ordnance-survey-mock:4011")
		payBaseUrl            = env.Get("GOVUK_PAY_BASE_URL", "http://pay-mock:4010")
		port                  = env.Get("APP_PORT", "8080")
		yotiClientSdkID       = env.Get("YOTI_CLIENT_SDK_ID", "")
		yotiScenarioID        = env.Get("YOTI_SCENARIO_ID", "")
		yotiSandbox           = env.Get("YOTI_SANDBOX", "") == "1"
		xrayEnabled           = env.Get("XRAY_ENABLED", "") == "1"
		rumConfig             = page.RumConfig{
			GuestRoleArn:      env.Get("AWS_RUM_GUEST_ROLE_ARN", ""),
			Endpoint:          env.Get("AWS_RUM_ENDPOINT", ""),
			ApplicationRegion: env.Get("AWS_RUM_APPLICATION_REGION", ""),
			IdentityPoolID:    env.Get("AWS_RUM_IDENTITY_POOL_ID", ""),
			ApplicationID:     env.Get("AWS_RUM_APPLICATION_ID", ""),
		}
	)

	paths := page.AppPaths{
		Auth:                                        "/auth",
		AuthRedirect:                                "/auth/redirect",
		AboutPayment:                                "/about-payment",
		CertificateProviderDetails:                  "/certificate-provider-details",
		CheckYourLpa:                                "/check-your-lpa",
		ChooseAttorneysAddress:                      "/choose-attorneys-address",
		ChooseAttorneys:                             "/choose-attorneys",
		ChooseAttorneysSummary:                      "/choose-attorneys-summary",
		ChoosePeopleToNotify:                        "/choose-people-to-notify",
		ChoosePeopleToNotifyAddress:                 "/choose-people-to-notify-address",
		ChoosePeopleToNotifySummary:                 "/choose-people-to-notify-summary",
		ChooseReplacementAttorneys:                  "/choose-replacement-attorneys",
		ChooseReplacementAttorneysAddress:           "/choose-replacement-attorneys-address",
		ChooseReplacementAttorneysSummary:           "/choose-replacement-attorneys-summary",
		CookiesConsent:                              "/cookies-consent",
		DoYouWantReplacementAttorneys:               "/want-replacement-attorneys",
		DoYouWantToNotifyPeople:                     "/do-you-want-to-notify-people",
		Dashboard:                                   "/dashboard",
		HealthCheck:                                 "/health-check",
		HowDoYouKnowYourCertificateProvider:         "/how-do-you-know-your-certificate-provider",
		HowLongHaveYouKnownCertificateProvider:      "/how-long-have-you-known-certificate-provider",
		HowShouldReplacementAttorneysMakeDecisions:  "/how-should-replacement-attorneys-make-decisions",
		HowShouldReplacementAttorneysStepIn:         "/how-should-replacement-attorneys-step-in",
		HowShouldAttorneysMakeDecisions:             "/how-should-attorneys-make-decisions",
		HowToSign:                                   "/how-to-sign",
		HowWouldYouLikeToBeContacted:                "/how-would-you-like-to-be-contacted",
		IdentityConfirmed:                           "/identity-confirmed",
		IdentityWithCouncilTaxBill:                  "/id/council-tax-bill",
		IdentityWithDrivingLicence:                  "/id/driving-licence",
		IdentityWithDwpAccount:                      "/id/dwp-account",
		IdentityWithGovernmentGatewayAccount:        "/id/government-gateway-account",
		IdentityWithOnlineBankAccount:               "/id/online-bank-account",
		IdentityWithPassport:                        "/id/passport",
		IdentityWithUtilityBill:                     "/id/utility-bill",
		IdentityWithYotiCallback:                    "/id/yoti/callback",
		IdentityWithYoti:                            "/id/yoti",
		LpaType:                                     "/lpa-type",
		PaymentConfirmation:                         "/payment-confirmation",
		ReadYourLpa:                                 "/read-your-lpa",
		RemoveAttorney:                              "/remove-attorney",
		RemovePersonToNotify:                        "/remove-person-to-notify",
		RemoveReplacementAttorney:                   "/remove-replacement-attorney",
		Restrictions:                                "/restrictions",
		Root:                                        "/",
		SelectYourIdentityOptions:                   "/select-your-identity-options",
		SigningConfirmation:                         "/signing-confirmation",
		Start:                                       "/start",
		TaskList:                                    "/task-list",
		TestingStart:                                "/testing-start",
		WhatHappensWhenSigning:                      "/what-happens-when-signing",
		WhenCanTheLpaBeUsed:                         "/when-can-the-lpa-be-used",
		WhoDoYouWantToBeCertificateProviderGuidance: "/who-do-you-want-to-be-certificate-provider-guidance",
		WhoIsTheLpaFor:                              "/who-is-the-lpa-for",
		YourAddress:                                 "/your-address",
		YourChosenIdentityOptions:                   "/your-chosen-identity-options",
		YourDetails:                                 "/your-details",
	}

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

	tmpls, err := template.Parse(webDir+"/template", templatefn.All)
	if err != nil {
		logger.Fatal(err)
	}

	bundle := localize.NewBundle("lang/en.json", "lang/cy.json")

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

	dynamoClient, err := dynamo.NewClient(cfg, dynamoTableLpas)
	if err != nil {
		logger.Fatal(err)
	}

	secretsClient, err := secrets.NewClient(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	sessionKeys, err := secretsClient.CookieSessionKeys(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	sessionStore := sessions.NewCookieStore(sessionKeys...)

	redirectURL := authRedirectBaseURL + paths.AuthRedirect

	signInClient, err := signin.Discover(ctx, logger, httpClient, secretsClient, issuer, clientID, redirectURL)
	if err != nil {
		logger.Fatal(err)
	}

	payApiKey, err := secretsClient.Secret(ctx, secrets.GovUkPay)
	if err != nil {
		logger.Fatal(err)
	}

	payClient := &pay.Client{
		BaseURL:    payBaseUrl,
		ApiKey:     payApiKey,
		HttpClient: httpClient,
	}

	yotiPrivateKey, err := secretsClient.SecretBytes(ctx, secrets.YotiPrivateKey)
	if err != nil {
		logger.Fatal(err)
	}

	yotiClient, err := identity.NewYotiClient(yotiClientSdkID, yotiPrivateKey)
	if err != nil {
		logger.Fatal(err)
	}
	if yotiSandbox {
		if err := yotiClient.SetupSandbox(); err != nil {
			logger.Fatal(err)
		}
	}

	osApiKey, err := secretsClient.Secret(ctx, secrets.OrdnanceSurvey)
	if err != nil {
		logger.Fatal(err)
	}

	addressClient := place.NewClient(ordnanceSurveyBaseUrl, osApiKey, httpClient)

	notifyApiKey, err := secretsClient.Secret(ctx, secrets.GovUkNotify)
	if err != nil {
		logger.Fatal(err)
	}

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, httpClient)
	if err != nil {
		logger.Fatal(err)
	}

	secureCookies := strings.HasPrefix(appPublicURL, "https:")

	mux := http.NewServeMux()
	mux.HandleFunc(paths.HealthCheck, func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"/robots.txt")
	})
	mux.Handle("/static/", http.StripPrefix("/static", handlers.CompressHandler(page.CacheControlHeaders(http.FileServer(http.Dir(webDir+"/static/"))))))
	mux.Handle(paths.AuthRedirect, page.AuthRedirect(logger, signInClient, sessionStore, secureCookies, paths))
	mux.Handle(paths.Auth, page.Login(logger, signInClient, sessionStore, secureCookies, random.String))
	mux.Handle(paths.CookiesConsent, page.CookieConsent(paths))
	mux.Handle("/cy/", http.StripPrefix("/cy", page.App(logger, bundle.For("cy"), page.Cy, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient, rumConfig, staticHash, paths)))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient, rumConfig, staticHash, paths))

	var handler http.Handler = mux
	if xrayEnabled {
		handler = telemetry.WrapHandler(mux)
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
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
