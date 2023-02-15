package app

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func App(
	logger page.Logger,
	localizer localize.Localizer,
	lang localize.Lang,
	tmpls template.Templates,
	sessionStore sesh.Store,
	dataStore page.DataStore,
	appPublicUrl string,
	payClient page.PayClient,
	yotiClient page.YotiClient,
	yotiScenarioID string,
	notifyClient page.NotifyClient,
	addressClient page.AddressClient,
	rumConfig page.RumConfig,
	staticHash string,
	paths page.AppPaths,
	oneLoginClient page.OneLoginClient,
) http.Handler {
	lpaStore := &lpaStore{dataStore: dataStore, randomInt: rand.Intn}

	rootMux := http.NewServeMux()

	rootMux.Handle(paths.TestingStart, page.TestingStart(sessionStore, lpaStore, random.String))
	rootMux.Handle(paths.Root, page.Root(paths))

	handleRoot := makeHandle(rootMux, logger, sessionStore)

	handleRoot(paths.Start, page.Guidance(tmpls.Get("start.gohtml"), paths.Auth, nil))

	certificateprovider.Register(
		rootMux,
		logger,
		tmpls,
		sessionStore,
		lpaStore,
		oneLoginClient,
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
