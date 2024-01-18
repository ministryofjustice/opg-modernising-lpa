package supporter

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	tmpls template.Templates,
	oneLoginClient OneLoginClient,
	sessionStore SessionStore,
	notFoundHandler page.Handler,
	errorHandler page.ErrorHandler,
) {
	paths := page.Paths.Supporter
	handleRoot := makeHandle(rootMux, errorHandler)

	handleRoot(paths.Login,
		page.Login(oneLoginClient, sessionStore, random.String, paths.LoginCallback))
	handleRoot(paths.LoginCallback,
		LoginCallback(oneLoginClient, sessionStore))

	supporterMux := http.NewServeMux()
	rootMux.Handle("/supporter/", http.StripPrefix("/supporter", supporterMux))

	handleSupporter := makeHandle(supporterMux, errorHandler)
	handleWithSupporter := makeSupporterHandle(supporterMux, sessionStore, errorHandler)

	handleSupporter(page.Paths.Root, notFoundHandler)

	handleWithSupporter(paths.YourOrganisation,
		TODO())
}

func makeHandle(mux *http.ServeMux, errorHandler page.ErrorHandler) func(page.Path, page.Handler) {
	return func(path page.Path, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeSupporterHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.SupporterPath, Handler) {
	return func(path page.SupporterPath, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			appData := page.AppDataFromContext(ctx)

			session, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Supporter.Start.Format(), http.StatusFound)
				return
			}

			appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				ctx = page.ContextWithSessionData(ctx, sessionData)
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
			}

			appData.Page = path.Format()

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
