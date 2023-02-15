package certificateprovider

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Register(
	rootMux *http.ServeMux,
	logger page.Logger,
	tmpls template.Templates,
	sessionStore sesh.Store,
	lpaStore page.LpaStore,
	oneLoginClient page.OneLoginClient,
	addressClient page.AddressClient,
) {
	handleRoot := makeHandle(rootMux, logger, sessionStore, None)

	handleRoot(page.Paths.CertificateProviderStart, None,
		Start(tmpls.Get("certificate_provider_start.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderLogin, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.CertificateProviderLoginCallback, None,
		LoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))
	handleRoot(page.Paths.CertificateProviderYourDetails, RequireSession,
		YourDetails(tmpls.Get("certificate_provider_your_details.gohtml"), lpaStore))
	handleRoot(page.Paths.CertificateProviderYourAddress, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))
	handleRoot(page.Paths.CertificateProviderReadTheLpa, RequireSession,
		page.Guidance(tmpls.Get("your_address.gohtml"), "/the-next-page", lpaStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, logger page.Logger, store sesh.Store, defaultOptions handleOpt) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := sesh.CertificateProvider(store, r)
				if err != nil {
					logger.Print(err)
					http.Redirect(w, r, page.Paths.CertificateProviderStart, http.StatusFound)
					return
				}

				appData.SessionID = session.DonorSessionID
				appData.LpaID = session.LpaID

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				str := fmt.Sprintf("Error rendering page for path '%s': %s", path, err.Error())

				logger.Print(str)
				http.Error(w, "Encountered an error", http.StatusInternalServerError)
			}
		})
	}
}
