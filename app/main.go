package main

import (
	"context"
	"fmt"
	html "html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/signin"
	"golang.org/x/exp/slices"
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
		ordnanceSurveyBaseUrl = env.Get("ORDNANCE_SURVEY_BASE_URL", "http://ordnance-survey-mock:4011")
		payBaseUrl            = env.Get("GOVUK_PAY_BASE_URL", "http://pay-mock:4010")
		port                  = env.Get("APP_PORT", "8080")
		yotiClientSdkID       = env.Get("YOTI_CLIENT_SDK_ID", "")
		yotiScenarioID        = env.Get("YOTI_SCENARIO_ID", "")
		yotiSandbox           = env.Get("YOTI_SANDBOX", "") == "1"
	)

	tmpls, err := template.Parse(webDir+"/template", map[string]interface{}{
		"isEnglish": func(lang page.Lang) bool {
			return lang == page.En
		},
		"isWelsh": func(lang page.Lang) bool {
			return lang == page.Cy
		},
		"input": func(top interface{}, name, label string, value interface{}, attrs ...interface{}) map[string]interface{} {
			field := map[string]interface{}{
				"top":   top,
				"name":  name,
				"label": label,
				"value": value,
			}

			if len(attrs)%2 != 0 {
				panic("must have even number of attrs")
			}

			for i := 0; i < len(attrs); i += 2 {
				field[attrs[i].(string)] = attrs[i+1]
			}

			return field
		},
		"items": func(top interface{}, name string, value interface{}, items ...interface{}) map[string]interface{} {
			return map[string]interface{}{
				"top":   top,
				"name":  name,
				"value": value,
				"items": items,
			}
		},
		"item": func(value, label string, attrs ...interface{}) map[string]interface{} {
			item := map[string]interface{}{
				"value": value,
				"label": label,
			}

			if len(attrs)%2 != 0 {
				panic("must have even number of attrs")
			}

			for i := 0; i < len(attrs); i += 2 {
				item[attrs[i].(string)] = attrs[i+1]
			}

			return item
		},
		"fieldID": func(name string, i int) string {
			if i == 0 {
				return name
			}

			return fmt.Sprintf("%s-%d", name, i+1)
		},
		"errorMessage": func(top interface{}, name string) map[string]interface{} {
			return map[string]interface{}{
				"top":  top,
				"name": name,
			}
		},
		"details": func(top interface{}, name, detail string) map[string]interface{} {
			return map[string]interface{}{
				"top":    top,
				"name":   name,
				"detail": detail,
			}
		},
		"inc": func(i int) int {
			return i + 1
		},
		"link": func(app page.AppData, path string) string {
			if app.Lang == page.Cy {
				return "/cy" + path
			}

			return path
		},
		"contains": func(needle string, list interface{}) bool {
			if slist, ok := list.([]string); ok {
				return slices.Contains(slist, needle)
			}

			if slist, ok := list.([]page.IdentityOption); ok {
				for _, item := range slist {
					if item.String() == needle {
						return true
					}
				}
			}

			return false
		},
		"tr": func(app page.AppData, messageID string) string {
			return app.Localizer.T(messageID)
		},
		"trFormat": func(app page.AppData, messageID string, args ...interface{}) string {
			if len(args)%2 != 0 {
				panic("must have even number of args")
			}

			data := map[string]interface{}{}
			for i := 0; i < len(args); i += 2 {
				data[args[i].(string)] = args[i+1]
			}

			return app.Localizer.Format(messageID, data)
		},
		"trFormatHtml": func(app page.AppData, messageID string, args ...interface{}) html.HTML {
			if len(args)%2 != 0 {
				panic("must have even number of args")
			}

			data := map[string]interface{}{}
			for i := 0; i < len(args); i += 2 {
				data[args[i].(string)] = args[i+1]
			}

			return html.HTML(app.Localizer.Format(messageID, data))
		},
		"trHtml": func(app page.AppData, messageID string) html.HTML {
			return html.HTML(app.Localizer.T(messageID))
		},
		"trCount": func(app page.AppData, messageID string, count int) string {
			return app.Localizer.Count(messageID, count)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"addDays": func(days int, t time.Time) time.Time {
			return t.AddDate(0, 0, days)
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}

			return t.Format("2 January 2006")
		},
		"formatDateTime": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}

			return t.Format("15:04:05, 2 January 2006")
		},
		"lowerFirst": func(s string) string {
			r, n := utf8.DecodeRuneInString(s)
			return string(unicode.ToLower(r)) + s[n:]
		},
	})
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

	secureCookies := strings.HasPrefix(appPublicURL, "https:")

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

	notifyClient, err := notify.New(notifyBaseURL, notifyApiKey, http.DefaultClient)
	if err != nil {
		logger.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(webDir+"/static/"))))
	mux.Handle(page.AuthRedirectPath, page.AuthRedirect(logger, signInClient, sessionStore, secureCookies))
	mux.Handle(page.AuthPath, page.Login(logger, signInClient, sessionStore, secureCookies, random.String))
	mux.Handle("/cookies-consent", page.CookieConsent())
	mux.Handle("/cy/", http.StripPrefix("/cy", page.App(logger, bundle.For("cy"), page.Cy, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient)))
	mux.Handle("/", page.App(logger, bundle.For("en"), page.En, tmpls, sessionStore, dynamoClient, appPublicURL, payClient, yotiClient, yotiScenarioID, notifyClient, addressClient))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
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
