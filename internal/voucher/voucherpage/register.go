// Package voucherpage provides the pages that a voucher interacts with.
package voucherpage

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error

type Template func(io.Writer, interface{}) error

type ErrorHandler func(http.ResponseWriter, *http.Request, error)

type Localizer interface {
	localize.Localizer
}

type LpaStoreResolvingService interface {
	Get(ctx context.Context) (*lpadata.Lpa, error)
}

type LpaStoreClient interface {
	SendDonorConfirmIdentity(ctx context.Context, donor *donordata.Provided) error
}

type DonorStore interface {
	GetAny(ctx context.Context) (*donordata.Provided, error)
	Put(ctx context.Context, donor *donordata.Provided) error
	FailVoucher(ctx context.Context, provided *donordata.Provided) error
}

type NotifyClient interface {
	EmailGreeting(lpa *lpadata.Lpa) string
	SendActorEmail(ctx context.Context, to notify.ToEmail, lpaUID string, email notify.Email) error
	SendActorSMS(ctx context.Context, to notify.ToMobile, lpaUID string, sms notify.SMS) error
}

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
	AuthCodeURL(state, nonce, locale string, confidenceLevel onelogin.ConfidenceLevel) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(userInfo onelogin.UserInfo) (identity.UserData, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode sharecodedata.Hashed) (sharecodedata.Link, error)
	Put(ctx context.Context, actorType actor.Type, shareCode sharecodedata.Hashed, data sharecodedata.Link) error
	Delete(ctx context.Context, shareCode sharecodedata.Link) error
}

type VoucherStore interface {
	Create(ctx context.Context, shareCode sharecodedata.Link, email string) (*voucherdata.Provided, error)
	Get(ctx context.Context) (*voucherdata.Provided, error)
	Put(ctx context.Context, provided *voucherdata.Provided) error
}

type DashboardStore interface {
	GetAll(ctx context.Context) (results dashboarddata.Results, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type ScheduledStore interface {
	Create(ctx context.Context, rows ...scheduled.Event) error
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
	lpaStoreResolvingService LpaStoreResolvingService,
	notifyClient NotifyClient,
	donorStore DonorStore,
	lpaStoreClient LpaStoreClient,
	scheduledStore ScheduledStore,
	donorStartURL string,
	localizer Localizer,
) {
	vouchFailed := makeVouchFailer(donorStore, notifyClient, donorStartURL, localizer)

	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.PathVoucherLogin, None,
		page.Login(oneLoginClient, sessionStore, random.AlphaNumeric, page.PathVoucherLoginCallback))
	handleRoot(page.PathVoucherLoginCallback, None,
		page.LoginCallback(logger, oneLoginClient, sessionStore, page.PathVoucherEnterReferenceNumber, dashboardStore, actor.TypeVoucher))
	handleRoot(page.PathVoucherEnterReferenceNumber, RequireSession,
		page.EnterAccessCode(tmpls.Get("enter_reference_number.gohtml"), shareCodeStore, sessionStore, lpaStoreResolvingService, actor.TypeVoucher,
			EnterAccessCode(voucherStore)))
	handleRoot(page.PathVoucherUnableToConfirmIdentity, None,
		page.Guidance(tmpls.Get("unable_to_confirm_identity.gohtml")))
	handleRoot(page.PathVoucherDonorDetailsDoNotMatch, None,
		page.Guidance(tmpls.Get("donor_details_do_not_match.gohtml")))
	handleRoot(page.PathYouCannotVouchForDonor, None,
		page.Guidance(tmpls.Get("you_cannot_vouch_for_donor.gohtml")))

	handleVoucher := makeVoucherHandle(rootMux, sessionStore, errorHandler, voucherStore)

	handleVoucher(voucher.PathTaskList, None,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStoreResolvingService))

	handleVoucher(voucher.PathConfirmYourName, None,
		ConfirmYourName(tmpls.Get("confirm_your_name.gohtml"), lpaStoreResolvingService, voucherStore))
	handleVoucher(voucher.PathYourName, None,
		YourName(tmpls.Get("your_name.gohtml"), lpaStoreResolvingService, voucherStore))
	handleVoucher(voucher.PathConfirmAllowedToVouch, None,
		ConfirmAllowedToVouch(tmpls.Get("confirm_allowed_to_vouch.gohtml"), lpaStoreResolvingService, voucherStore, vouchFailed))

	handleVoucher(voucher.PathVerifyDonorDetails, None,
		VerifyDonorDetails(tmpls.Get("verify_donor_details.gohtml"), lpaStoreResolvingService, voucherStore, vouchFailed, donorStore))

	handleVoucher(voucher.PathConfirmYourIdentity, None,
		ConfirmYourIdentity(tmpls.Get("confirm_your_identity.gohtml"), voucherStore, lpaStoreResolvingService))
	handleVoucher(voucher.PathHowWillYouConfirmYourIdentity, None,
		HowWillYouConfirmYourIdentity(tmpls.Get("how_will_you_confirm_your_identity.gohtml"), voucherStore))
	handleVoucher(voucher.PathCompletingYourIdentityConfirmation, None,
		CompletingYourIdentityConfirmation(tmpls.Get("completing_your_identity_confirmation.gohtml"), lpaStoreResolvingService))
	handleVoucher(voucher.PathIdentityWithOneLogin, None,
		IdentityWithOneLogin(oneLoginClient, sessionStore, random.AlphaNumeric))
	handleVoucher(voucher.PathIdentityWithOneLoginCallback, None,
		IdentityWithOneLoginCallback(oneLoginClient, sessionStore, voucherStore, lpaStoreResolvingService, vouchFailed))
	handleVoucher(voucher.PathOneLoginIdentityDetails, None,
		Guidance(tmpls.Get("one_login_identity_details.gohtml"), lpaStoreResolvingService))

	handleVoucher(voucher.PathSignTheDeclaration, None,
		YourDeclaration(tmpls.Get("your_declaration.gohtml"), lpaStoreResolvingService, voucherStore, donorStore, notifyClient, lpaStoreClient, scheduledStore, time.Now, donorStartURL, localizer))
	handleVoucher(voucher.PathThankYou, None,
		Guidance(tmpls.Get("thank_you.gohtml"), lpaStoreResolvingService))
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
			appData.ActorType = actor.TypeVoucher

			if opt&RequireSession != 0 {
				session, err := store.Login(r)
				if err != nil {
					http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
					return
				}

				appData.SessionID = session.SessionID()
				appData.HasLpas = session.HasLPAs

				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func makeVoucherHandle(mux *http.ServeMux, store SessionStore, errorHandler page.ErrorHandler, voucherStore VoucherStore) func(voucher.Path, handleOpt, Handler) {
	return func(path voucher.Path, opt handleOpt, h Handler) {
		mux.HandleFunc(path.String(), func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := appcontext.DataFromContext(ctx)
			appData.CanGoBack = opt&CanGoBack != 0
			appData.ActorType = actor.TypeVoucher
			appData.LpaID = r.PathValue("id")
			appData.Page = path.Format(appData.LpaID)

			session, err := store.Login(r)
			if err != nil {
				http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
				return
			}

			appData.SessionID = session.SessionID()
			appData.HasLpas = session.HasLPAs

			sessionData, err := appcontext.SessionFromContext(ctx)
			if err == nil {
				sessionData.SessionID = appData.SessionID
				sessionData.LpaID = appData.LpaID
				ctx = appcontext.ContextWithSession(ctx, sessionData)
			} else {
				ctx = appcontext.ContextWithSession(ctx, &appcontext.Session{SessionID: appData.SessionID, LpaID: appData.LpaID})
			}

			provided, err := voucherStore.Get(ctx)
			if err != nil {
				errorHandler(w, r, err)
				return
			}

			if !path.CanGoTo(provided) {
				voucher.PathTaskList.Redirect(w, r, appData, provided.LpaID)
				return
			}

			if err := h(appData, w, r.WithContext(appcontext.ContextWithData(ctx, appData)), provided); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}
