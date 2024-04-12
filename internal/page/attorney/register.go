package attorney

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpastore.Lpa, error)
}

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, details *actor.AttorneyProvidedDetails) error

type Template func(io.Writer, interface{}) error

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type SessionStore interface {
	Login(r *http.Request) (*sesh.LoginSession, error)
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (actor.ShareCodeData, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error
	Delete(ctx context.Context, shareCode actor.ShareCodeData) error
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
}

type AttorneyStore interface {
	Create(ctx context.Context, sessionID string, attorneyUID actoruid.UID, isReplacement, isTrustCorporation bool) (*actor.AttorneyProvidedDetails, error)
	Get(ctx context.Context) (*actor.AttorneyProvidedDetails, error)
	GetAny(ctx context.Context) ([]*actor.AttorneyProvidedDetails, error)
	Put(ctx context.Context, attorney *actor.AttorneyProvidedDetails) error
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	SendAttorney(context.Context, *lpastore.Lpa, *actor.AttorneyProvidedDetails) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	commonTmpls, tmpls template.Templates,
	sessionStore SessionStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	notFoundHandler page.Handler,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
	lpaStoreResolvingService LpaStoreResolvingService,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Login, None,
		page.Login(oneLoginClient, sessionStore, random.String, page.Paths.Attorney.LoginCallback))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.Attorney.EnterReferenceNumber, dashboardStore, actor.TypeAttorney))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, attorneyStore))

	handleAttorney := makeAttorneyHandle(rootMux, sessionStore, errorHandler, attorneyStore)

	handleAttorney(page.Paths.Attorney.CodeOfConduct, RequireAttorney,
		Guidance(tmpls.Get("code_of_conduct.gohtml"), lpaStoreResolvingService))
	handleAttorney(page.Paths.Attorney.TaskList, RequireAttorney,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStoreResolvingService, certificateProviderStore))
	handleAttorney(page.Paths.Attorney.MobileNumber, RequireAttorney,
		MobileNumber(tmpls.Get("mobile_number.gohtml"), attorneyStore))
	handleAttorney(page.Paths.Attorney.YourPreferredLanguage, RequireAttorney,
		YourPreferredLanguage(commonTmpls.Get("your_preferred_language.gohtml"), attorneyStore, lpaStoreResolvingService))
	handleAttorney(page.Paths.Attorney.ConfirmYourDetails, RequireAttorney,
		ConfirmYourDetails(tmpls.Get("confirm_your_details.gohtml"), attorneyStore, lpaStoreResolvingService))
	handleAttorney(page.Paths.Attorney.ReadTheLpa, RequireAttorney,
		ReadTheLpa(tmpls.Get("read_the_lpa.gohtml"), lpaStoreResolvingService, attorneyStore))
	handleAttorney(page.Paths.Attorney.RightsAndResponsibilities, RequireAttorney,
		Guidance(tmpls.Get("legal_rights_and_responsibilities.gohtml"), nil))
	handleAttorney(page.Paths.Attorney.WhatHappensWhenYouSign, RequireAttorney,
		Guidance(tmpls.Get("what_happens_when_you_sign.gohtml"), lpaStoreResolvingService))
	handleAttorney(page.Paths.Attorney.Sign, RequireAttorney,
		Sign(tmpls.Get("sign.gohtml"), lpaStoreResolvingService, certificateProviderStore, attorneyStore, lpaStoreClient, time.Now))
	handleAttorney(page.Paths.Attorney.WouldLikeSecondSignatory, RequireAttorney,
		WouldLikeSecondSignatory(tmpls.Get("would_like_second_signatory.gohtml"), attorneyStore, lpaStoreResolvingService, lpaStoreClient))
	handleAttorney(page.Paths.Attorney.WhatHappensNext, RequireAttorney,
		Guidance(tmpls.Get("what_happens_next.gohtml"), lpaStoreResolvingService))
	handleAttorney(page.Paths.Attorney.Progress, RequireAttorney,
		Progress(tmpls.Get("progress.gohtml"), lpaStoreResolvingService))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	RequireAttorney
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := store.Login(r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = session.SessionID()
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeAttorneyHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, attorneyStore AttorneyStore) func(page.AttorneyPath, handleOpt, Handler) {
	return func(path page.AttorneyPath, opt handleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = page.ContextWithSessionData(ctx, sessionData)
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			attorney, err := attorneyStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			appData.Page = path.Format(appData.LpaID)
			appData.AttorneyUID = attorney.UID
			if attorney.IsTrustCorporation && attorney.IsReplacement {
				appData.ActorType = actor.TypeReplacementTrustCorporation
			} else if attorney.IsTrustCorporation {
				appData.ActorType = actor.TypeTrustCorporation
			} else if attorney.IsReplacement {
				appData.ActorType = actor.TypeReplacementAttorney
			} else {
				appData.ActorType = actor.TypeAttorney
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData)), attorney); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
