// Package attorneypage provides the pages that an attorney or trust corporation
// interacts with.
package attorneypage

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

type Localizer interface {
	page.Localizer
}

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpadata.Lpa, error)
}

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request, details *attorneydata.Provided) error

type Template func(io.Writer, interface{}) error

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
	LpaData(r *http.Request) (*sesh.LpaDataSession, error)
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *sesh.LpaDataSession) error
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (sharecodedata.Link, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data sharecodedata.Link) error
	Delete(ctx context.Context, shareCode sharecodedata.Link) error
}

type AttorneyStore interface {
	Create(ctx context.Context, shareCode sharecodedata.Link, email string) (*attorneydata.Provided, error)
	Get(ctx context.Context) (*attorneydata.Provided, error)
	Put(ctx context.Context, attorney *attorneydata.Provided) error
	Delete(ctx context.Context) error
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type DashboardStore interface {
	GetAll(ctx context.Context) (results dashboarddata.Results, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	SendAttorney(context.Context, *lpadata.Lpa, *attorneydata.Provided) error
	SendAttorneyOptOut(ctx context.Context, lpaUID string, attorneyUID actoruid.UID, actorType actor.Type) error
}

type NotifyClient interface {
	EmailGreeting(lpa *lpadata.Lpa) string
	SendActorEmail(ctx context.Context, lang localize.Lang, to, lpaUID string, email notify.Email) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
	sessionStore SessionStore,
	attorneyStore AttorneyStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
	lpaStoreResolvingService LpaStoreResolvingService,
	notifyClient NotifyClient,
	appPublicURL string,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.PathAttorneyLogin, None,
		page.Login(oneLoginClient, sessionStore, random.String, page.PathAttorneyLoginCallback))
	handleRoot(page.PathAttorneyLoginCallback, None,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.PathAttorneyEnterReferenceNumber, dashboardStore, actor.TypeAttorney))
	handleRoot(page.PathAttorneyEnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, attorneyStore))
	handleRoot(page.PathAttorneyEnterReferenceNumberOptOut, None,
		EnterReferenceNumberOptOut(tmpls.Get("enter_reference_number_opt_out.gohtml"), shareCodeStore, sessionStore))
	handleRoot(page.PathAttorneyConfirmDontWantToBeAttorneyLoggedOut, None,
		ConfirmDontWantToBeAttorneyLoggedOut(tmpls.Get("confirm_dont_want_to_be_attorney.gohtml"), shareCodeStore, lpaStoreResolvingService, sessionStore, notifyClient, appPublicURL, lpaStoreClient))
	handleRoot(page.PathAttorneyYouHaveDecidedNotToBeAttorney, None,
		page.Guidance(tmpls.Get("you_have_decided_not_to_be_attorney.gohtml")))

	handleAttorney := makeAttorneyHandle(rootMux, sessionStore, errorHandler, attorneyStore)

	handleAttorney(attorney.PathCodeOfConduct, None,
		Guidance(tmpls.Get("code_of_conduct.gohtml"), lpaStoreResolvingService))
	handleAttorney(attorney.PathTaskList, None,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStoreResolvingService))
	handleAttorney(attorney.PathPhoneNumber, None,
		PhoneNumber(tmpls.Get("phone_number.gohtml"), attorneyStore))
	handleAttorney(attorney.PathYourPreferredLanguage, CanGoBack,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), attorneyStore, lpaStoreResolvingService))
	handleAttorney(attorney.PathConfirmYourDetails, None,
		ConfirmYourDetails(tmpls.Get("confirm_your_details.gohtml"), attorneyStore, lpaStoreResolvingService))
	handleAttorney(attorney.PathReadTheLpa, None,
		ReadTheLpa(tmpls.Get("read_the_lpa.gohtml"), lpaStoreResolvingService, attorneyStore))
	handleAttorney(attorney.PathRightsAndResponsibilities, None,
		Guidance(tmpls.Get("legal_rights_and_responsibilities.gohtml"), nil))
	handleAttorney(attorney.PathWhatHappensWhenYouSign, CanGoBack,
		Guidance(tmpls.Get("what_happens_when_you_sign.gohtml"), lpaStoreResolvingService))
	handleAttorney(attorney.PathSign, CanGoBack,
		Sign(tmpls.Get("sign.gohtml"), lpaStoreResolvingService, attorneyStore, lpaStoreClient, time.Now))
	handleAttorney(attorney.PathWouldLikeSecondSignatory, None,
		WouldLikeSecondSignatory(tmpls.Get("would_like_second_signatory.gohtml"), attorneyStore, lpaStoreResolvingService, lpaStoreClient))
	handleAttorney(attorney.PathWhatHappensNext, None,
		Guidance(tmpls.Get("what_happens_next.gohtml"), lpaStoreResolvingService))
	handleAttorney(attorney.PathProgress, None,
		Progress(tmpls.Get("progress.gohtml"), lpaStoreResolvingService))

	handleAttorney(attorney.PathConfirmDontWantToBeAttorney, CanGoBack,
		ConfirmDontWantToBeAttorney(tmpls.Get("confirm_dont_want_to_be_attorney.gohtml"), lpaStoreResolvingService, attorneyStore, notifyClient, appPublicURL, lpaStoreClient))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.Page = path.Format()
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := store.Login(r)
				if err != nil {
					http.Redirect(w, r, page.PathAttorneyStart.Format(), http.StatusFound)
					return
				}

				appData.SessionID = session.SessionID()
				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeAttorneyHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, attorneyStore AttorneyStore) func(attorney.Path, handleOpt, Handler) {
	return func(path attorney.Path, opt handleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.PathAttorneyStart.Format(), http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()

			sessionData, err := appcontext.SessionFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = appcontext.ContextWithSession(ctx, sessionData)
			} else {
				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			provided, err := attorneyStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !attorney.CanGoTo(provided, r.URL.String()) {
				attorney.PathTaskList.Redirect(w, r, appData, provided.LpaID)
				return
			}

			appData.Page = path.Format(appData.LpaID)
			appData.AttorneyUID = provided.UID
			if provided.IsTrustCorporation && provided.IsReplacement {
				appData.ActorType = actor.TypeReplacementTrustCorporation
			} else if provided.IsTrustCorporation {
				appData.ActorType = actor.TypeTrustCorporation
			} else if provided.IsReplacement {
				appData.ActorType = actor.TypeReplacementAttorney
			} else {
				appData.ActorType = actor.TypeAttorney
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData)), provided); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
