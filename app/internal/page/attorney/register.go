package attorney

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	GetAny(context.Context) (*page.Lpa, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeStore --structname mockShareCodeStore
type ShareCodeStore interface {
	Get(context.Context, actor.Type, string) (actor.ShareCodeData, error)
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

//go:generate mockery --testonly --inpackage --name CertificateProviderStore --structname mockCertificateProviderStore
type CertificateProviderStore interface {
	GetAny(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
}

//go:generate mockery --testonly --inpackage --name AttorneyStore --structname mockAttorneyStore
type AttorneyStore interface {
	Create(context.Context, bool) (*actor.AttorneyProvidedDetails, error)
	Get(context.Context) (*actor.AttorneyProvidedDetails, error)
	Put(context.Context, *actor.AttorneyProvidedDetails) error
}

//go:generate mockery --testonly --inpackage --name AddressClient --structname mockAddressClient
type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	oneLoginClient OneLoginClient,
	addressClient AddressClient,
	shareCodeStore ShareCodeStore,
	errorHandler page.ErrorHandler,
	notifyClient NotifyClient,
	notFoundHandler page.Handler,
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("attorney_enter_reference_number.gohtml"), shareCodeStore, sessionStore, attorneyStore))

	attorneyMux := http.NewServeMux()
	rootMux.Handle("/attorney/", routeToPrefix("/attorney/", attorneyMux, notFoundHandler))
	handleAttorney := makeHandle(attorneyMux, sessionStore, errorHandler)

	handleAttorney(page.Paths.Attorney.CodeOfConduct, RequireSession,
		Guidance(tmpls.Get("attorney_code_of_conduct.gohtml"), donorStore))
	handleAttorney(page.Paths.Attorney.TaskList, RequireSession,
		TaskList(tmpls.Get("attorney_task_list.gohtml"), donorStore, certificateProviderStore, attorneyStore))
	handleAttorney(page.Paths.Attorney.CheckYourName, RequireSession,
		CheckYourName(tmpls.Get("attorney_check_your_name.gohtml"), donorStore, attorneyStore, notifyClient))
	handleAttorney(page.Paths.Attorney.DateOfBirth, RequireSession,
		DateOfBirth(tmpls.Get("attorney_date_of_birth.gohtml"), attorneyStore))
	handleAttorney(page.Paths.Attorney.MobileNumber, RequireSession,
		MobileNumber(tmpls.Get("attorney_mobile_number.gohtml"), attorneyStore))
	handleAttorney(page.Paths.Attorney.YourAddress, RequireSession,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, attorneyStore))
	handleAttorney(page.Paths.Attorney.ReadTheLpa, RequireSession,
		ReadTheLpa(tmpls.Get("attorney_read_the_lpa.gohtml"), donorStore, attorneyStore))
	handleAttorney(page.Paths.Attorney.RightsAndResponsibilities, RequireSession,
		page.Guidance(tmpls.Get("attorney_legal_rights_and_responsibilities.gohtml")))
	handleAttorney(page.Paths.Attorney.WhatHappensWhenYouSign, RequireSession,
		Guidance(tmpls.Get("attorney_what_happens_when_you_sign.gohtml"), donorStore))
	handleAttorney(page.Paths.Attorney.Sign, RequireSession,
		Sign(tmpls.Get("attorney_sign.gohtml"), donorStore, certificateProviderStore, attorneyStore))
	handleAttorney(page.Paths.Attorney.WhatHappensNext, RequireSession,
		Guidance(tmpls.Get("attorney_what_happens_next.gohtml"), donorStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, errorHandler page.ErrorHandler) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.ServiceName = "beAnAttorney"
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := sesh.Attorney(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))
				appData.AttorneyID = session.AttorneyID
				appData.LpaID = session.LpaID

				if session.IsReplacementAttorney {
					appData.ActorType = actor.TypeReplacementAttorney
				} else {
					appData.ActorType = actor.TypeAttorney
				}

				ctx = page.ContextWithSessionData(ctx, &page.SessionData{
					SessionID: appData.SessionID,
					LpaID:     appData.LpaID,
				})
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func routeToPrefix(prefix string, mux http.Handler, notFoundHandler page.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 4)
		if len(parts) != 4 {
			notFoundHandler(page.AppDataFromContext(r.Context()), w, r)
			return
		}

		id, path := parts[2], "/"+parts[3]

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = path
		if len(r.URL.RawPath) > len(prefix)+len(id) {
			r2.URL.RawPath = r.URL.RawPath[len(prefix)+len(id):]
		}

		mux.ServeHTTP(w, r2.WithContext(page.ContextWithSessionData(r2.Context(), &page.SessionData{LpaID: id})))
	}
}
