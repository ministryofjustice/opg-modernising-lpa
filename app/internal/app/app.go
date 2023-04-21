package app

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
)

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	Get(ctx context.Context, pk, sk string, v interface{}) error
	Put(context.Context, string, string, interface{}) error
	GetOneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
	GetAllByGsi(ctx context.Context, gsi, sk string, v interface{}) error
}

func App(
	logger *logging.Logger,
	localizer page.Localizer,
	lang localize.Lang,
	tmpls template.Templates,
	sessionStore sesh.Store,
	dataStore DataStore,
	appPublicUrl string,
	payClient *pay.Client,
	yotiClient *identity.YotiClient,
	notifyClient *notify.Client,
	addressClient *place.Client,
	rumConfig page.RumConfig,
	staticHash string,
	paths page.AppPaths,
	oneLoginClient *onelogin.Client,
) http.Handler {
	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}
	certificateProviderStore := &certificateProviderStore{dataStore: dataStore, now: time.Now}

	shareCodeSender := page.NewShareCodeSender(dataStore, notifyClient, appPublicUrl, random.String)

	errorHandler := page.Error(tmpls.Get("error-500.gohtml"), logger)
	notFoundHandler := page.Root(tmpls.Get("error-404.gohtml"), logger)

	rootMux := http.NewServeMux()

	rootMux.Handle(paths.TestingStart, page.TestingStart(sessionStore, lpaStore, random.String, shareCodeSender, localizer, certificateProviderStore))

	handleRoot := makeHandle(rootMux, errorHandler)

	handleRoot(paths.Root, notFoundHandler)
	handleRoot(paths.Fixtures, page.Fixtures(tmpls.Get("fixtures.gohtml")))
	handleRoot(paths.YourLegalRightsAndResponsibilities, page.Guidance(tmpls.Get("your_legal_rights_and_responsibilities.gohtml"), nil))

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		lpaStore,
		oneLoginClient,
		dataStore,
		addressClient,
		errorHandler,
		yotiClient,
		notifyClient,
		certificateProviderStore,
	)

	attorney.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		lpaStore,
		oneLoginClient,
		dataStore,
		errorHandler,
		notifyClient,
	)

	donor.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		lpaStore,
		oneLoginClient,
		addressClient,
		appPublicUrl,
		payClient,
		yotiClient,
		notifyClient,
		shareCodeSender,
		errorHandler,
		notFoundHandler,
		certificateProviderStore,
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

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler) func(string, page.Handler) {
	return func(path string, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path

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
