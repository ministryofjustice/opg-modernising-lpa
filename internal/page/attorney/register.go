package attorney

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type Handler func(data page.AppData, w http.ResponseWriter, r *http.Request, details *actor.AttorneyProvidedDetails) error

type Template func(io.Writer, interface{}) error

type Logger interface {
	Print(v ...interface{})
}

type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type DonorStore interface {
	GetAny(context.Context) (*actor.DonorProvidedDetails, error)
}

type ShareCodeStore interface {
	Get(context.Context, actor.Type, string) (actor.ShareCodeData, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error
}

type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
}

type AttorneyStore interface {
	Create(context.Context, string, string, bool, bool) (*actor.AttorneyProvidedDetails, error)
	Get(context.Context) (*actor.AttorneyProvidedDetails, error)
	GetAny(context.Context) ([]*actor.AttorneyProvidedDetails, error)
	Put(context.Context, *actor.AttorneyProvidedDetails) error
}

type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaStoreClient interface {
	SendAttorney(context.Context, *actor.DonorProvidedDetails, *actor.AttorneyProvidedDetails) error
}

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	notFoundHandler page.Handler,
	dashboardStore DashboardStore,
	lpaStoreClient LpaStoreClient,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Login, None,
		page.Login(oneLoginClient, sessionStore, random.String, page.Paths.Attorney.LoginCallback))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		page.LoginCallback(oneLoginClient, sessionStore, page.Paths.Attorney.EnterReferenceNumber, dashboardStore, actor.TypeAttorney))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("attorney_enter_reference_number.gohtml"), shareCodeStore, sessionStore, attorneyStore))

	attorneyMux := http.NewServeMux()
	rootMux.Handle("/attorney/", page.RouteToPrefix("/attorney/", attorneyMux, notFoundHandler))
	handleAttorney := makeAttorneyHandle(attorneyMux, sessionStore, errorHandler, attorneyStore)

	handleAttorney(page.Paths.Attorney.CodeOfConduct, RequireAttorney,
		Guidance(tmpls.Get("attorney_code_of_conduct.gohtml"), donorStore))
	handleAttorney(page.Paths.Attorney.TaskList, RequireAttorney,
		TaskList(tmpls.Get("attorney_task_list.gohtml"), donorStore, certificateProviderStore))
	handleAttorney(page.Paths.Attorney.MobileNumber, RequireAttorney,
		MobileNumber(tmpls.Get("attorney_mobile_number.gohtml"), attorneyStore))
	handleAttorney(page.Paths.Attorney.YourPreferredLanguage, RequireAttorney,
		YourPreferredLanguage(tmpls.Get("your_preferred_language.gohtml"), attorneyStore, donorStore))
	handleAttorney(page.Paths.Attorney.ConfirmYourDetails, RequireAttorney,
		ConfirmYourDetails(tmpls.Get("attorney_confirm_your_details.gohtml"), attorneyStore, donorStore))
	handleAttorney(page.Paths.Attorney.ReadTheLpa, RequireAttorney,
		ReadTheLpa(tmpls.Get("attorney_read_the_lpa.gohtml"), donorStore, attorneyStore))
	handleAttorney(page.Paths.Attorney.RightsAndResponsibilities, RequireAttorney,
		Guidance(tmpls.Get("attorney_legal_rights_and_responsibilities.gohtml"), nil))
	handleAttorney(page.Paths.Attorney.WhatHappensWhenYouSign, RequireAttorney,
		Guidance(tmpls.Get("attorney_what_happens_when_you_sign.gohtml"), donorStore))
	handleAttorney(page.Paths.Attorney.Sign, RequireAttorney,
		Sign(tmpls.Get("attorney_sign.gohtml"), donorStore, certificateProviderStore, attorneyStore, lpaStoreClient, time.Now))
	handleAttorney(page.Paths.Attorney.WouldLikeSecondSignatory, RequireAttorney,
		WouldLikeSecondSignatory(tmpls.Get("attorney_would_like_second_signatory.gohtml"), attorneyStore, donorStore, lpaStoreClient))
	handleAttorney(page.Paths.Attorney.WhatHappensNext, RequireAttorney,
		Guidance(tmpls.Get("attorney_what_happens_next.gohtml"), donorStore))
	handleAttorney(page.Paths.Attorney.Progress, RequireAttorney,
		Progress(tmpls.Get("attorney_progress.gohtml"), attorneyStore, donorStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	RequireAttorney
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(page.Path, handleOpt, page.Handler) {
	return func(path page.Path, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path.Format()
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := sesh.Login(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeAttorneyHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler, attorneyStore AttorneyStore) func(page.AttorneyPath, handleOpt, Handler) {
	return func(path page.AttorneyPath, opt handleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0

			session, err := sesh.Login(store, r)
			if err != nil {
				http.Redirect(w, r, page.Paths.Attorney.Start.Format(), http.StatusFound)
				return
			}

			appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))

			sessionData, err := page.SessionDataFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				ctx = page.ContextWithSessionData(ctx, sessionData)

				appData.LpaID = sessionData.LpaID
			} else {
				ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			attorney, err := attorneyStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			appData.Page = path.Format(appData.LpaID)
			appData.AttorneyID = attorney.ID
			if attorney.IsReplacement {
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
