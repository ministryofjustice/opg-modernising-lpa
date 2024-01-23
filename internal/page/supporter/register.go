package supporter

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type OrganisationStore interface {
	Create(context.Context, string) error
	CreateMemberInvite(context.Context, *actor.Organisation, string, string) error
	Get(context.Context) (*actor.Organisation, error)
}

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

type NotifyClient interface {
	SendEmail(context.Context, string, notify.Email) (string, error)
}

type Template func(io.Writer, interface{}) error

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	tmpls template.Templates,
	oneLoginClient OneLoginClient,
	sessionStore SessionStore,
	organisationStore OrganisationStore,
	notFoundHandler page.Handler,
	errorHandler page.ErrorHandler,
	notifyClient NotifyClient,
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

	handleWithSupporter(paths.EnterOrganisationName,
		EnterOrganisationName(tmpls.Get("enter_organisation_name.gohtml"), organisationStore))
	handleWithSupporter(paths.OrganisationCreated,
		OrganisationCreated(tmpls.Get("organisation_created.gohtml"), organisationStore))
	handleWithSupporter(paths.Dashboard,
		TODO())
	handleWithSupporter(paths.InviteMember,
		InviteMember(tmpls.Get("invite_member.gohtml"), organisationStore, notifyClient, random.String))
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

			appData.Page = path.Format()
			appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))

			ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
