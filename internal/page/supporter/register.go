package supporter

import (
	"context"
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
	Create(ctx context.Context, name string) (*actor.Organisation, error)
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, email, code string) error
	Get(ctx context.Context) (*actor.Organisation, error)
	CreateLPA(ctx context.Context, organisationID string) (*actor.DonorProvidedDetails, error)
	GetMember(ctx context.Context) (*actor.Member, error)
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
	SendEmail(context context.Context, to string, email notify.Email) error
}

type DonorStore interface {
	Get(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Latest(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Put(ctx context.Context, donor *actor.DonorProvidedDetails) error
	Delete(ctx context.Context) error
}

type Template func(io.Writer, interface{}) error

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	supporterTmpls, donorTmpls template.Templates,
	oneLoginClient OneLoginClient,
	sessionStore SessionStore,
	organisationStore OrganisationStore,
	notFoundHandler page.Handler,
	errorHandler page.ErrorHandler,
	notifyClient NotifyClient,
	donorStore DonorStore,
) {
	supporterPaths := page.Paths.Supporter
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(supporterPaths.Login, page.None,
		page.Login(oneLoginClient, sessionStore, random.String, supporterPaths.LoginCallback))
	handleRoot(supporterPaths.LoginCallback, page.None,
		LoginCallback(oneLoginClient, sessionStore, organisationStore))
	handleRoot(supporterPaths.EnterOrganisationName, page.RequireSession,
		EnterOrganisationName(supporterTmpls.Get("enter_organisation_name.gohtml"), organisationStore, sessionStore))

	supporterMux := http.NewServeMux()
	rootMux.Handle("/supporter/", http.StripPrefix("/supporter", supporterMux))

	handleSupporter := makeHandle(supporterMux, sessionStore, errorHandler)
	handleWithSupporter := makeSupporterHandle(supporterMux, sessionStore, errorHandler, organisationStore)

	handleSupporter(page.Paths.Root, page.None, notFoundHandler)

	handleWithSupporter(supporterPaths.OrganisationCreated,
		OrganisationCreated(supporterTmpls.Get("organisation_created.gohtml")))
	handleWithSupporter(supporterPaths.Dashboard,
		Dashboard(supporterTmpls.Get("dashboard.gohtml"), organisationStore))
	handleWithSupporter(supporterPaths.InviteMember,
		InviteMember(supporterTmpls.Get("invite_member.gohtml"), organisationStore, notifyClient, random.String))
	handleWithSupporter(supporterPaths.InviteMemberConfirmation,
		Guidance(supporterTmpls.Get("invite_member_confirmation.gohtml")))
}

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.Path, page.HandleOpt, page.Handler) {
	return func(path page.Path, opt page.HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.IsSupporter = true

			if opt&page.RequireSession != 0 {
				session, err := sesh.Login(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Supporter.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = session.SessionID()

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeSupporterHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler, organisationStore OrganisationStore) func(page.SupporterPath, Handler) {
	return func(path page.SupporterPath, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			loginSession, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Supporter.Start.Format(), http.StatusFound)
				return
			}

			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.SessionID = loginSession.SessionID()

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				ctx = page.ContextWithSessionData(ctx, sessionData)
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
			}

			organisation, err := organisationStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			appData.Page = path.Format()
			appData.IsSupporter = true
			appData.OrganisationID = organisation.ID
			appData.OrganisationName = organisation.Name

			ctx = page.ContextWithAppData(page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, OrganisationID: organisation.ID}), appData)

			if err := h(appData, w, r.WithContext(ctx), organisation); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
