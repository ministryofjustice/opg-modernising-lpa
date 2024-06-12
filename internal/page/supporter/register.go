package supporter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpastore.Lpa, error)
}

type OrganisationStore interface {
	Create(ctx context.Context, member *actor.Member, name string) (*actor.Organisation, error)
	CreateLPA(ctx context.Context) (*actor.DonorProvidedDetails, error)
	Get(ctx context.Context) (*actor.Organisation, error)
	Put(ctx context.Context, organisation *actor.Organisation) error
	SoftDelete(ctx context.Context, organisation *actor.Organisation) error
}

type MemberStore interface {
	Create(ctx context.Context, firstNames, lastName string) (*actor.Member, error)
	CreateFromInvite(ctx context.Context, invite *actor.MemberInvite) error
	CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error
	DeleteMemberInvite(ctx context.Context, organisationID, email string) error
	Get(ctx context.Context) (*actor.Member, error)
	GetAny(ctx context.Context) (*actor.Member, error)
	GetAll(ctx context.Context) ([]*actor.Member, error)
	GetByID(ctx context.Context, memberID string) (*actor.Member, error)
	InvitedMember(ctx context.Context) (*actor.MemberInvite, error)
	InvitedMembers(ctx context.Context) ([]*actor.MemberInvite, error)
	InvitedMembersByEmail(ctx context.Context) ([]*actor.MemberInvite, error)
	Put(ctx context.Context, member *actor.Member) error
}

type DonorStore interface {
	DeleteDonorAccess(ctx context.Context, shareCodeData actor.ShareCodeData) error
	Get(ctx context.Context) (*actor.DonorProvidedDetails, error)
	GetByKeys(ctx context.Context, keys []dynamo.Keys) ([]actor.DonorProvidedDetails, error)
	Put(ctx context.Context, donor *actor.DonorProvidedDetails) error
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
}

type Localizer interface {
	page.Localizer
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type SessionStore interface {
	ClearLogin(r *http.Request, w http.ResponseWriter) error
	Login(r *http.Request) (*sesh.LoginSession, error)
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type NotifyClient interface {
	SendEmail(ctx context.Context, to string, email notify.Email) error
}

type ShareCodeStore interface {
	PutDonor(ctx context.Context, shareCode string, data actor.ShareCodeData) error
	GetDonor(ctx context.Context) (actor.ShareCodeData, error)
	Delete(ctx context.Context, data actor.ShareCodeData) error
}

type Template func(w io.Writer, data interface{}) error

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, member *actor.Member) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type ProgressTracker interface {
	Progress(lpa *lpastore.Lpa) page.Progress
}

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
	searchClient *search.Client,
	donorStore DonorStore,
	shareCodeStore ShareCodeStore,
	progressTracker ProgressTracker,
	lpaStoreResolvingService LpaStoreResolvingService,
) {
	paths := page.Paths.Supporter
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(paths.Start, None,
		page.Guidance(tmpls.Get("start.gohtml")))
	handleRoot(paths.SigningInAdvice, None,
		page.Guidance(tmpls.Get("signing_in_advice.gohtml")))
	handleRoot(paths.Login, None,
		page.Login(oneLoginClient, sessionStore, random.String, paths.LoginCallback))
	handleRoot(paths.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore, organisationStore, time.Now, memberStore))
	handleRoot(paths.EnterYourName, RequireSession,
		EnterYourName(tmpls.Get("enter_your_name.gohtml"), memberStore))
	handleRoot(paths.EnterOrganisationName, RequireSession,
		EnterOrganisationName(tmpls.Get("enter_organisation_name.gohtml"), organisationStore, memberStore, sessionStore))
	handleRoot(paths.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), memberStore, sessionStore))
	handleRoot(paths.InviteExpired, RequireSession,
		page.Guidance(tmpls.Get("invite_expired.gohtml")))
	handleRoot(paths.OrganisationDeleted, None,
		page.Guidance(tmpls.Get("organisation_deleted.gohtml")))

	handleWithSupporter := makeSupporterHandle(rootMux, sessionStore, errorHandler, organisationStore, memberStore, tmpls.Get("suspended.gohtml"))

	handleWithSupporter(paths.OrganisationCreated, None,
		Guidance(tmpls.Get("organisation_created.gohtml")))
	handleWithSupporter(paths.Dashboard, None,
		Dashboard(tmpls.Get("dashboard.gohtml"), donorStore, searchClient))
	handleWithSupporter(paths.ConfirmDonorCanInteractOnline, None,
		ConfirmDonorCanInteractOnline(tmpls.Get("confirm_donor_can_interact_online.gohtml"), organisationStore))
	handleWithSupporter(paths.ContactOPGForPaperForms, None,
		Guidance(tmpls.Get("contact_opg_for_paper_forms.gohtml")))
	handleWithSupporter(paths.ViewLPA, None,
		ViewLPA(tmpls.Get("view_lpa.gohtml"), lpaStoreResolvingService, progressTracker))

	handleWithSupporter(paths.OrganisationDetails, RequireAdmin,
		Guidance(tmpls.Get("organisation_details.gohtml")))
	handleWithSupporter(paths.EditOrganisationName, RequireAdmin,
		EditOrganisationName(tmpls.Get("edit_organisation_name.gohtml"), organisationStore))
	handleWithSupporter(paths.ManageTeamMembers, RequireAdmin,
		ManageTeamMembers(tmpls.Get("manage_team_members.gohtml"), memberStore, random.String, notifyClient, appPublicURL))
	handleWithSupporter(paths.InviteMember, CanGoBack|RequireAdmin,
		InviteMember(tmpls.Get("invite_member.gohtml"), memberStore, notifyClient, random.String, appPublicURL))
	handleWithSupporter(paths.DeleteOrganisation, CanGoBack,
		DeleteOrganisation(tmpls.Get("delete_organisation.gohtml"), organisationStore, sessionStore, searchClient))
	handleWithSupporter(paths.EditMember, CanGoBack,
		EditMember(tmpls.Get("edit_member.gohtml"), memberStore))

	handleWithSupporter(paths.DonorAccess, CanGoBack,
		DonorAccess(tmpls.Get("donor_access.gohtml"), donorStore, shareCodeStore, notifyClient, appPublicURL, random.String))
}

type HandleOpt byte

const (
	None HandleOpt = 1 << iota
	RequireSession
	RequireAdmin
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler) func(page.Path, HandleOpt, page.Handler) {
	return func(path page.Path, opt HandleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.CanToggleWelsh = false
			appData.SupporterData = &page.SupporterData{}

			if opt&RequireSession != 0 {
				session, err := store.Login(r)
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

type suspendedData struct {
	App              page.AppData
	Errors           validation.List
	OrganisationName string
}

type SupporterPath interface {
	String() string
	IsManageOrganisation() bool
}

func makeSupporterHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, organisationStore OrganisationStore, memberStore MemberStore, suspendedTmpl template.Template) func(SupporterPath, HandleOpt, Handler) {
	return func(path SupporterPath, opt HandleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			loginSession, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Supporter.Start.Format(), http.StatusFound)
				return
			}

			appData := page.AppDataFromContext(r.Context())
			appData.SessionID = loginSession.SessionID()
			appData.CanGoBack = opt&CanGoBack != 0
			appData.CanToggleWelsh = false

			appData.SupporterData = &page.SupporterData{
				IsManageOrganisation: path.IsManageOrganisation(),
			}

			appData.LoginSessionEmail = loginSession.Email

			switch v := path.(type) {
			case page.SupporterPath:
				appData.Page = v.Format()
			case page.SupporterLpaPath:
				appData.LpaID = r.PathValue("id")
				appData.Page = v.Format(appData.LpaID)
			default:
				panic("non-supporter path registered")
			}

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
				LpaID:          appData.LpaID,
			})

			member, err := memberStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if opt&RequireAdmin != 0 && !member.Permission.IsAdmin() {
				errorHandler(w, r, errors.New("permission denied"))
				return
			}

			if member.Status.IsSuspended() {
				if err := suspendedTmpl(w, &suspendedData{
					App:              appData,
					OrganisationName: organisation.Name,
				}); err != nil {
					errorHandler(w, r, err)
				}

				return
			}

			appData.SupporterData.OrganisationName = organisation.Name
			appData.SupporterData.Permission = member.Permission
			appData.SupporterData.LoggedInSupporterID = member.ID

			ctx = page.ContextWithAppData(ctx, appData)

			if err := h(appData, w, r.WithContext(ctx), organisation, member); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
