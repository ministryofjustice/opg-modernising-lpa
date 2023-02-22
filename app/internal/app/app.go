package app

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/logging"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	GetAll(context.Context, string, interface{}) error
	Get(context.Context, string, string, interface{}) error
	Put(context.Context, string, string, interface{}) error
}

func App(
	logger *logging.Logger,
	localizer localize.Localizer,
	lang localize.Lang,
	tmpls template.Templates,
	sessionStore sesh.Store,
	dataStore DataStore,
	appPublicUrl string,
	payClient *pay.Client,
	yotiClient *identity.YotiClient,
	yotiScenarioID string,
	notifyClient *notify.Client,
	addressClient *place.Client,
	rumConfig page.RumConfig,
	staticHash string,
	paths page.AppPaths,
	oneLoginClient *onelogin.Client,
) http.Handler {
	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}
	shareCodeSender := page.NewShareCodeSender(dataStore, notifyClient, appPublicUrl, random.String)

	rootMux := http.NewServeMux()

	rootMux.Handle(paths.TestingStart, page.TestingStart(sessionStore, lpaStore, random.String, dataStore, shareCodeSender))
	rootMux.Handle(paths.Root, page.Root(paths))

	handleRoot := makeHandle(rootMux, logger, sessionStore)

	handleRoot(paths.Start, page.Guidance(tmpls.Get("start.gohtml"), nil))
	handleRoot(paths.Fixtures, page.Fixtures(tmpls.Get("fixtures.gohtml")))

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		lpaStore,
		oneLoginClient,
		dataStore,
		addressClient,
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
		yotiScenarioID,
		notifyClient,
		shareCodeSender,
	)

	return withAppData(page.ValidateCsrf(rootMux, sessionStore, random.String), localizer, lang, rumConfig, staticHash)
}

func withAppData(next http.Handler, localizer localize.Localizer, lang localize.Lang, rumConfig page.RumConfig, staticHash string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		appData := page.AppDataFromContext(ctx)
		appData.Query = queryString(r)
		appData.Localizer = localizer
		appData.Lang = lang
		appData.RumConfig = rumConfig
		appData.StaticHash = staticHash
		appData.Paths = page.Paths
		appData.Localizer.ShowTranslationKeys = r.FormValue("showTranslationKeys") == "1"

		_, cookieErr := r.Cookie("cookies-consent")
		appData.CookieConsentSet = cookieErr != http.ErrNoCookie

		next.ServeHTTP(w, r.WithContext(page.ContextWithAppData(ctx, appData)))
	}
}

func makeHandle(mux *http.ServeMux, logger page.Logger, store sesh.Store) func(string, page.Handler) {
	return func(path string, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				str := fmt.Sprintf("Error rendering page for path '%s': %s", path, err.Error())

				logger.Print(str)
				http.Error(w, "Encountered an error", http.StatusInternalServerError)
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
