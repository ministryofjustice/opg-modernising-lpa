package supporter

import (
	"context"
	"io"
	"net/http"
	"time"

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
	AllLPAs(ctx context.Context) ([]actor.DonorProvidedDetails, error)
	Create(ctx context.Context, name string) (*actor.Organisation, error)
	CreateLPA(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Get(ctx context.Context) (*actor.Organisation, error)
	Put(ctx context.Context, organisation *actor.Organisation) error
}

type MemberStore interface {
	Create(ctx context.Context, invite *actor.MemberInvite) error
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error
	InvitedMember(ctx context.Context) (*actor.MemberInvite, error)
	InvitedMembers(ctx context.Context) ([]*actor.MemberInvite, error)
	Get(ctx context.Context) (*actor.Member, error)
	GetByID(ctx context.Context, memberID string) (*actor.Member, error)
	GetAll(ctx context.Context) ([]*actor.Member, error)
	Put(ctx context.Context, member *actor.Member) error
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

type Template func(io.Writer, interface{}) error

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	tmpls template.Templates,
	oneLoginClient OneLoginClient,
	sessionStore SessionStore,
	organisationStore OrganisationStore,
	errorHandler page.ErrorHandler,
	notifyClient NotifyClient,
	appPublicURL string,
	memberStore MemberStore,
) {
	paths := page.Paths.Supporter
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(paths.SigningInAdvice, page.None,
		page.Guidance(tmpls.Get("signing_in_advice.gohtml")))
	handleRoot(paths.Login, page.None,
		page.Login(oneLoginClient, sessionStore, random.String, paths.LoginCallback))
	handleRoot(paths.LoginCallback, page.None,
		LoginCallback(oneLoginClient, sessionStore, organisationStore, time.Now, memberStore))
	handleRoot(paths.EnterOrganisationName, page.RequireSession,
		EnterOrganisationName(tmpls.Get("enter_organisation_name.gohtml"), organisationStore, sessionStore))
	handleRoot(paths.EnterReferenceNumber, page.RequireSession,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), memberStore, sessionStore))
	handleRoot(paths.InviteExpired, page.RequireSession,
		page.Guidance(tmpls.Get("invite_expired.gohtml")))

	handleWithSupporter := makeSupporterHandle(rootMux, sessionStore, errorHandler, organisationStore, memberStore)

	handleWithSupporter(paths.OrganisationCreated, page.None,
		Guidance(tmpls.Get("organisation_created.gohtml")))
	handleWithSupporter(paths.Dashboard, page.None,
		Dashboard(tmpls.Get("dashboard.gohtml"), organisationStore))
	handleWithSupporter(paths.ConfirmDonorCanInteractOnline, page.None,
		ConfirmDonorCanInteractOnline(tmpls.Get("confirm_donor_can_interact_online.gohtml"), organisationStore))
	handleWithSupporter(paths.ContactOPGForPaperForms, page.None,
		Guidance(tmpls.Get("contact_opg_for_paper_forms.gohtml")))

	handleWithSupporter(paths.OrganisationDetails, page.None,
		Guidance(tmpls.Get("organisation_details.gohtml")))
	handleWithSupporter(paths.EditOrganisationName, page.None,
		EditOrganisationName(tmpls.Get("edit_organisation_name.gohtml"), organisationStore))
	handleWithSupporter(paths.ManageTeamMembers, page.None,
		ManageTeamMembers(tmpls.Get("manage_team_members.gohtml"), memberStore))
	handleWithSupporter(paths.InviteMember, page.CanGoBack,
		InviteMember(tmpls.Get("invite_member.gohtml"), memberStore, notifyClient, random.String, appPublicURL))
	handleWithSupporter(paths.EditMember, page.CanGoBack,
		EditMember(tmpls.Get("edit_team_member.gohtml"), memberStore))
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

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, Email: session.Email})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeSupporterHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler, organisationStore OrganisationStore, memberStore MemberStore) func(page.SupporterPath, page.HandleOpt, Handler) {
	return func(path page.SupporterPath, opt page.HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			loginSession, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Supporter.Start.Format(), http.StatusFound)
				return
			}

			appData := page.AppDataFromContext(r.Context())
			appData.SessionID = loginSession.SessionID()
			appData.CanGoBack = opt&page.CanGoBack != 0

			sessionData, err := page.SessionDataFromContext(r.Context())

			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.OrganisationID = loginSession.OrganisationID
			} else {
				sessionData = &page.SessionData{
					SessionID: appData.SessionID,
					Email:     loginSession.Email,
				}

				if loginSession.OrganisationID != "" {
					sessionData.OrganisationID = loginSession.OrganisationID
				}
			}

			organisation, err := organisationStore.Get(page.ContextWithSessionData(r.Context(), sessionData))
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
				SessionID:      appData.SessionID,
				Email:          loginSession.Email,
				OrganisationID: organisation.ID,
			})

			member, err := memberStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			appData.Page = path.Format()
			appData.IsSupporter = true
			appData.OrganisationName = organisation.Name
			appData.IsManageOrganisation = path.IsManageOrganisation()
			appData.LoginSessionEmail = loginSession.Email
			appData.Permission = member.Permission
			appData.LoggedInSupporterID = member.ID

			ctx = page.ContextWithAppData(ctx, appData)

			if err := h(appData, w, r.WithContext(ctx), organisation); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
