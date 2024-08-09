// Package voucherpage provides the pages that a voucher interacts with.
package voucherpage

import (
	"context"
	"io"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request) error

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

type VoucherStore interface {
	Create(ctx context.Context, shareCode sharecodedata.Link, email string) (any, error)
}

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []page.LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	voucherStore VoucherStore,
	oneLoginClient OneLoginClient,
	shareCodeStore ShareCodeStore,
	dashboardStore DashboardStore,
	errorHandler page.ErrorHandler,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.PathVoucherLogin, None,
		page.Login(oneLoginClient, sessionStore, random.String, page.PathVoucherLoginCallback))
	handleRoot(page.PathVoucherLoginCallback, None,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.PathVoucherEnterReferenceNumber, dashboardStore, actor.TypeVoucher))
	handleRoot(page.PathVoucherEnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, voucherStore))

	handleVoucher := makeVoucherHandle(rootMux, sessionStore, errorHandler)

	handleVoucher(voucher.PathTaskList, None,
		TaskList(tmpls.Get("task_list.gohtml")))
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
					http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
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

func makeVoucherHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler) func(voucher.Path, handleOpt, Handler) {
	return func(path voucher.Path, opt handleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0
			appData.LpaID = r.PathValue("id")

			session, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
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

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
