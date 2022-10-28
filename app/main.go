package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/felixge/httpsnoop"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/templatefn"
	"go.opentelemetry.io/contrib/detectors/aws/ecs"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
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
	)

	if xrayEnabled {
		logger.Print("creating tracer, hopefully")
		shutdown, err := makeTracer(ctx, xrayEnabled)
		logger.Print("make tracer call finished")

		if err != nil {
			logger.Fatal(err)
		}
		defer func(ctx context.Context) {
			if err := shutdown(ctx); err != nil {
				logger.Fatal(fmt.Errorf("shutdown err: %w", err))
			}
		}(ctx)
	}

	tmpls, err := template.Parse(webDir+"/template", templatefn.All)
	if err != nil {
		logger.Fatal(err)
	}

	bundle := localize.NewBundle("lang/en.json", "lang/cy.json")

	config := &aws.Config{}
	if len(awsBaseURL) > 0 {
		config.Endpoint = aws.String(awsBaseURL)
	}

	sess, err := session.NewSession(config)
	if err != nil {
		logger.Fatal(fmt.Errorf("error initialising new AWS session: %w", err))
	}

	dynamoClient, err := dynamo.NewClient(sess, dynamoTableLpas)
	if err != nil {
		logger.Fatal(err)
	}

	secretsClient, err := secrets.NewClient(sess)
	if err != nil {
		logger.Fatal(err)
	}

	sessionKeys, err := secretsClient.CookieSessionKeys()
	if err != nil {
		logger.Fatal(err)
	}

	sessionStore := sessions.NewCookieStore(sessionKeys...)

	redirectURL := authRedirectBaseURL + page.AuthRedirectPath

	signInClient, err := signin.Discover(ctx, logger, http.DefaultClient, secretsClient, issuer, clientID, redirectURL)
	if err != nil {
		logger.Fatal(err)
	}

	payApiKey, err := secretsClient.Secret(secrets.GovUkPay)
	if err != nil {
		logger.Fatal(err)
	}

	payClient := &pay.Client{
		BaseURL:    payBaseUrl,
		ApiKey:     payApiKey,
		HttpClient: http.DefaultClient,
	}

	yotiPrivateKey, err := secretsClient.SecretBytes(secrets.YotiPrivateKey)
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

	osApiKey, err := secretsClient.Secret(secrets.OrdnanceSurvey)
	if err != nil {
		logger.Fatal(err)
	}

	addressClient := place.NewClient(ordnanceSurveyBaseUrl, osApiKey, http.DefaultClient)

	notifyApiKey, err := secretsClient.Secret(secrets.GovUkNotify)
	if err != nil {
		logger.Fatal(err)
	}

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, notifyApiKey, http.DefaultClient)
	if err != nil {
		logger.Fatal(err)
	}

	secureCookies := strings.HasPrefix(appPublicURL, "https:")

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(webDir+"/static/"))))
	mux.Handle(page.AuthRedirectPath, page.AuthRedirect(logger, signInClient, sessionStore, secureCookies))
	mux.Handle(page.AuthPath, page.Login(logger, signInClient, sessionStore, secureCookies, random.String))
	mux.Handle("/cookies-consent", page.CookieConsent())
	mux.Handle("/cy/", http.StripPrefix("/cy", page.App(logger, bundle.For("cy"), page.Cy, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient)))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient))

	var handler http.Handler = mux
	if xrayEnabled {
		handler = traceHandler(logger, mux)
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

func makeTracer(ctx context.Context, secure bool) (func(context.Context) error, error) {
	secureOption := otlptracegrpc.WithInsecure()
	// if secure {
	//	secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	// }

	traceExporter, err := otlptracegrpc.New(ctx,
		secureOption,
		otlptracegrpc.WithEndpoint("0.0.0.0:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()))
	if err != nil {
		return nil, fmt.Errorf("failed to create new OTLP trace exporter: %w", err)
	}

	ecsResourceDetector := ecs.NewResourceDetector()
	resource, err := ecsResourceDetector.Detect(ctx)
	if err != nil {
		return nil, err
	}

	// client := otlptracehttp.NewClient(otlptracehttp.WithInsecure())
	// traceExporter, err := otlptrace.New(ctx, client)
	// if err != nil {
	//	return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
	// }

	idg := xray.NewIDGenerator()

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})

	return traceExporter.Shutdown, nil
}

func traceHandler(logger *logging.Logger, handler http.Handler) http.HandlerFunc {
	tracer := otel.GetTracerProvider().Tracer("mlpab")

	return func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path
		isWelsh := false
		if strings.HasPrefix(r.URL.Path, "/cy/") {
			route = route[3:]
			isWelsh = true
		}

		target := r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}

		ctx, span := tracer.Start(r.Context(), route,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attribute.Bool("mlpab.welsh", isWelsh)),
			trace.WithAttributes(semconv.HTTPTargetKey.String(target)),
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(r)...),
			trace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest("mlpab", route, r)...),
		)
		defer span.End()

		m := httpsnoop.CaptureMetrics(handler, w, r.WithContext(ctx))

		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(m.Code)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(m.Code, trace.SpanKindServer))
		span.SetAttributes(semconv.HTTPResponseContentLengthKey.Int64(m.Written))
	}
}
