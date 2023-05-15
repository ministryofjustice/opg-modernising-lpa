package attorney

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
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
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	Get(context.Context) (*page.Lpa, error)
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
	Get(ctx context.Context) (*actor.CertificateProviderProvidedDetails, error)
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
) {
	handleRoot := makeHandle(rootMux, sessionStore, errorHandler)

	handleRoot(page.Paths.Attorney.Start, None,
		page.Guidance(tmpls.Get("attorney_start.gohtml")))
	handleRoot(page.Paths.Attorney.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.Attorney.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.Attorney.EnterReferenceNumber, RequireSession,
		EnterReferenceNumber(tmpls.Get("attorney_enter_reference_number.gohtml"), shareCodeStore, sessionStore, attorneyStore))
	handleRoot(page.Paths.Attorney.CodeOfConduct, RequireLpa,
		donor.Guidance(tmpls.Get("attorney_code_of_conduct.gohtml"), donorStore))
	handleRoot(page.Paths.Attorney.TaskList, RequireLpa,
		TaskList(tmpls.Get("attorney_task_list.gohtml"), donorStore, certificateProviderStore, attorneyStore))
	handleRoot(page.Paths.Attorney.CheckYourName, RequireLpa,
		CheckYourName(tmpls.Get("attorney_check_your_name.gohtml"), donorStore, attorneyStore, notifyClient))
	handleRoot(page.Paths.Attorney.DateOfBirth, RequireLpa,
		DateOfBirth(tmpls.Get("attorney_date_of_birth.gohtml"), attorneyStore))
	handleRoot(page.Paths.Attorney.MobileNumber, RequireLpa,
		MobileNumber(tmpls.Get("attorney_mobile_number.gohtml"), attorneyStore))
	handleRoot(page.Paths.Attorney.YourAddress, RequireLpa,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, attorneyStore))
	handleRoot(page.Paths.Attorney.ReadTheLpa, RequireLpa,
		ReadTheLpa(tmpls.Get("attorney_read_the_lpa.gohtml"), donorStore, attorneyStore))
	handleRoot(page.Paths.Attorney.RightsAndResponsibilities, RequireLpa,
		page.Guidance(tmpls.Get("attorney_legal_rights_and_responsibilities.gohtml")))
	handleRoot(page.Paths.Attorney.WhatHappensWhenYouSign, RequireLpa,
		donor.Guidance(tmpls.Get("attorney_what_happens_when_you_sign.gohtml"), donorStore))
	handleRoot(page.Paths.Attorney.Sign, RequireLpa,
		Sign(tmpls.Get("attorney_sign.gohtml"), donorStore, certificateProviderStore, attorneyStore))
	handleRoot(page.Paths.Attorney.WhatHappensNext, RequireLpa,
		donor.Guidance(tmpls.Get("attorney_what_happens_next.gohtml"), donorStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	RequireLpa
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
				if _, err := sesh.Attorney(store, r); err != nil {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}
			}

			if opt&RequireLpa != 0 {
				session, err := sesh.Attorney(store, r)
				if err != nil || session.LpaID == "" || session.AttorneyID == "" {
					http.Redirect(w, r, page.Paths.Attorney.Start, http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))
				appData.LpaID = session.LpaID
				appData.AttorneyID = session.AttorneyID

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
