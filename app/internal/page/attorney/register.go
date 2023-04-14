package attorney

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	oneLoginClient OneLoginClient,
	errorHandler page.ErrorHandler,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Start, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml"), nil))
	handleRoot(page.Paths.Attorney.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		page.Guidance(tmpls.Get("attorney_start.gohtml"), nil))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ServiceName = "beAnAttorney"
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				_, err := sesh.Attorney(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
