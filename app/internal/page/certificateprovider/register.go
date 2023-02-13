package certificateprovider

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Register(
	rootMux *http.ServeMux,
	logger page.Logger,
	sessionStore sesh.Store,
	localizer localize.Localizer,
	lang localize.Lang,
	rumConfig page.RumConfig,
	staticHash string,
	lpaStore page.LpaStore,
	tmpls template.Templates,
	oneLoginClient page.OneLoginClient,
) {
	handleRoot := makeHandle(rootMux, logger, sessionStore, localizer, lang, rumConfig, staticHash, None)

	handleRoot(page.Paths.CertificateProviderStart, None,
		Start(tmpls.Get("certificate_provider_start.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderLogin, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProviderLoginCallback, None,
		LoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))
	handleRoot(page.Paths.CertificateProviderYourDetails, RequireSession,
		page.Guidance(tmpls.Get("certificate_provider_your_details.gohtml"), "", lpaStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, logger page.Logger, store sesh.Store, localizer localize.Localizer, lang localize.Lang, rumConfig page.RumConfig, staticHash string, defaultOptions handleOpt) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppData{
				Page:       path,
				Query:      queryString(r),
				Localizer:  localizer,
				Lang:       lang,
				CanGoBack:  opt&CanGoBack != 0,
				RumConfig:  rumConfig,
				StaticHash: staticHash,
				Paths:      page.Paths,
				CsrfToken:  page.CsrfFromContext(ctx),
			}

			if opt&RequireSession != 0 {
				session, err := sesh.GetCertificateProviderSession(store, r)
				if err != nil {
					logger.Print(err)
					http.Redirect(w, r, page.Paths.Start, http.StatusFound)
					return
				}

				appData.SessionID = session.DonorSessionID
				appData.LpaID = session.LpaID

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			_, cookieErr := r.Cookie("cookies-consent")
			appData.CookieConsentSet = cookieErr != http.ErrNoCookie
			appData.Localizer.ShowTranslationKeys = r.FormValue("showTranslationKeys") == "1"

			if err := h(appData, w, r.WithContext(ctx)); err != nil {
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
