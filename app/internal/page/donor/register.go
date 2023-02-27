package donor

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
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

//go:generate mockery --testonly --inpackage --name LpaStore --structname mockLpaStore
type LpaStore interface {
	Create(context.Context) (*page.Lpa, error)
	GetAll(context.Context) ([]*page.Lpa, error)
	Get(context.Context) (*page.Lpa, error)
	Put(context.Context, *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name PayClient --structname mockPayClient
type PayClient interface {
	CreatePayment(body pay.CreatePaymentBody) (pay.CreatePaymentResponse, error)
	GetPayment(paymentId string) (pay.GetPaymentResponse, error)
}

//go:generate mockery --testonly --inpackage --name AddressClient --structname mockAddressClient
type AddressClient interface {
	LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeSender --structname mockShareCodeSender
type ShareCodeSender interface {
	Send(ctx context.Context, template notify.TemplateId, appData page.AppData, identity bool, lpa *page.Lpa) error
}

//go:generate mockery --testonly --inpackage --name YotiClient --structname mockYotiClient
type YotiClient interface {
	IsTest() bool
	SdkID() string
	User(string) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

func Register(
	rootMux *http.ServeMux,
	logger Logger,
	tmpls template.Templates,
	sessionStore SessionStore,
	lpaStore LpaStore,
	oneLoginClient OneLoginClient,
	addressClient AddressClient,
	appPublicUrl string,
	payClient PayClient,
	yotiClient YotiClient,
	yotiScenarioID string,
	notifyClient NotifyClient,
	shareCodeSender ShareCodeSender,
	errorHandler page.ErrorHandler,
	notFoundHandler page.Handler,
) {
	handleRoot := makeHandle(rootMux, sessionStore, None, errorHandler)

	handleRoot(page.Paths.Start, None,
		page.Guidance(tmpls.Get("start.gohtml"), nil))
	handleRoot(page.Paths.Login, None,
		Login(logger, oneLoginClient, sessionStore, random.String))
	handleRoot(page.Paths.LoginCallback, None,
		LoginCallback(oneLoginClient, sessionStore))
	handleRoot(page.Paths.Dashboard, RequireSession,
		Dashboard(tmpls.Get("dashboard.gohtml"), lpaStore))

	lpaMux := http.NewServeMux()

	rootMux.Handle("/lpa/", routeToLpa(lpaMux, notFoundHandler))

	handleLpa := makeHandle(lpaMux, sessionStore, RequireSession, errorHandler)

	handleLpa(page.Paths.Root, None, notFoundHandler)
	handleLpa(page.Paths.YourDetails, None,
		YourDetails(tmpls.Get("your_details.gohtml"), lpaStore, sessionStore))
	handleLpa(page.Paths.YourAddress, None,
		YourAddress(logger, tmpls.Get("your_address.gohtml"), addressClient, lpaStore))
	handleLpa(page.Paths.LpaType, None,
		LpaType(tmpls.Get("lpa_type.gohtml"), lpaStore))
	handleLpa(page.Paths.WhoIsTheLpaFor, None,
		WhoIsTheLpaFor(tmpls.Get("who_is_the_lpa_for.gohtml"), lpaStore))

	handleLpa(page.Paths.TaskList, None,
		TaskList(tmpls.Get("task_list.gohtml"), lpaStore))

	handleLpa(page.Paths.ChooseAttorneys, CanGoBack,
		ChooseAttorneys(tmpls.Get("choose_attorneys.gohtml"), lpaStore, random.String))
	handleLpa(page.Paths.ChooseAttorneysAddress, CanGoBack,
		ChooseAttorneysAddress(logger, tmpls.Get("choose_attorneys_address.gohtml"), addressClient, lpaStore))
	handleLpa(page.Paths.ChooseAttorneysSummary, CanGoBack,
		ChooseAttorneysSummary(logger, tmpls.Get("choose_attorneys_summary.gohtml"), lpaStore))
	handleLpa(page.Paths.RemoveAttorney, CanGoBack,
		RemoveAttorney(logger, tmpls.Get("remove_attorney.gohtml"), lpaStore))
	handleLpa(page.Paths.HowShouldAttorneysMakeDecisions, CanGoBack,
		HowShouldAttorneysMakeDecisions(tmpls.Get("how_should_attorneys_make_decisions.gohtml"), lpaStore))

	handleLpa(page.Paths.DoYouWantReplacementAttorneys, CanGoBack,
		WantReplacementAttorneys(tmpls.Get("do_you_want_replacement_attorneys.gohtml"), lpaStore))
	handleLpa(page.Paths.ChooseReplacementAttorneys, CanGoBack,
		ChooseReplacementAttorneys(tmpls.Get("choose_replacement_attorneys.gohtml"), lpaStore, random.String))
	handleLpa(page.Paths.ChooseReplacementAttorneysAddress, CanGoBack,
		ChooseReplacementAttorneysAddress(logger, tmpls.Get("choose_replacement_attorneys_address.gohtml"), addressClient, lpaStore))
	handleLpa(page.Paths.ChooseReplacementAttorneysSummary, CanGoBack,
		ChooseReplacementAttorneysSummary(logger, tmpls.Get("choose_replacement_attorneys_summary.gohtml"), lpaStore))
	handleLpa(page.Paths.RemoveReplacementAttorney, CanGoBack,
		RemoveReplacementAttorney(logger, tmpls.Get("remove_replacement_attorney.gohtml"), lpaStore))
	handleLpa(page.Paths.HowShouldReplacementAttorneysStepIn, CanGoBack,
		HowShouldReplacementAttorneysStepIn(tmpls.Get("how_should_replacement_attorneys_step_in.gohtml"), lpaStore))
	handleLpa(page.Paths.HowShouldReplacementAttorneysMakeDecisions, CanGoBack,
		HowShouldReplacementAttorneysMakeDecisions(tmpls.Get("how_should_replacement_attorneys_make_decisions.gohtml"), lpaStore))

	handleLpa(page.Paths.WhenCanTheLpaBeUsed, CanGoBack,
		WhenCanTheLpaBeUsed(tmpls.Get("when_can_the_lpa_be_used.gohtml"), lpaStore))
	handleLpa(page.Paths.Restrictions, CanGoBack,
		Restrictions(tmpls.Get("restrictions.gohtml"), lpaStore))
	handleLpa(page.Paths.WhoDoYouWantToBeCertificateProviderGuidance, CanGoBack,
		WhoDoYouWantToBeCertificateProviderGuidance(tmpls.Get("who_do_you_want_to_be_certificate_provider_guidance.gohtml"), lpaStore))
	handleLpa(page.Paths.CertificateProviderDetails, CanGoBack,
		CertificateProviderDetails(tmpls.Get("certificate_provider_details.gohtml"), lpaStore))
	handleLpa(page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole, CanGoBack,
		HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpls.Get("how_would_certificate_provider_prefer_to_carry_out_their_role.gohtml"), lpaStore))
	handleLpa(page.Paths.CertificateProviderAddress, CanGoBack,
		CertificateProviderAddress(logger, tmpls.Get("certificate_provider_address.gohtml"), addressClient, lpaStore))
	handleLpa(page.Paths.HowDoYouKnowYourCertificateProvider, CanGoBack,
		HowDoYouKnowYourCertificateProvider(tmpls.Get("how_do_you_know_your_certificate_provider.gohtml"), lpaStore))
	handleLpa(page.Paths.HowLongHaveYouKnownCertificateProvider, CanGoBack,
		HowLongHaveYouKnownCertificateProvider(tmpls.Get("how_long_have_you_known_certificate_provider.gohtml"), lpaStore))

	handleLpa(page.Paths.DoYouWantToNotifyPeople, CanGoBack,
		DoYouWantToNotifyPeople(tmpls.Get("do_you_want_to_notify_people.gohtml"), lpaStore))
	handleLpa(page.Paths.ChoosePeopleToNotify, CanGoBack,
		ChoosePeopleToNotify(tmpls.Get("choose_people_to_notify.gohtml"), lpaStore, random.String))
	handleLpa(page.Paths.ChoosePeopleToNotifyAddress, CanGoBack,
		ChoosePeopleToNotifyAddress(logger, tmpls.Get("choose_people_to_notify_address.gohtml"), addressClient, lpaStore))
	handleLpa(page.Paths.ChoosePeopleToNotifySummary, CanGoBack,
		ChoosePeopleToNotifySummary(logger, tmpls.Get("choose_people_to_notify_summary.gohtml"), lpaStore))
	handleLpa(page.Paths.RemovePersonToNotify, CanGoBack,
		RemovePersonToNotify(logger, tmpls.Get("remove_person_to_notify.gohtml"), lpaStore))

	handleLpa(page.Paths.CheckYourLpa, CanGoBack,
		CheckYourLpa(tmpls.Get("check_your_lpa.gohtml"), lpaStore))

	handleLpa(page.Paths.AboutPayment, CanGoBack,
		AboutPayment(logger, tmpls.Get("about_payment.gohtml"), sessionStore, payClient, appPublicUrl, random.String, lpaStore))
	handleLpa(page.Paths.PaymentConfirmation, CanGoBack,
		PaymentConfirmation(logger, tmpls.Get("payment_confirmation.gohtml"), payClient, lpaStore, sessionStore, shareCodeSender))

	handleLpa(page.Paths.HowToConfirmYourIdentityAndSign, CanGoBack,
		page.Guidance(tmpls.Get("how_to_confirm_your_identity_and_sign.gohtml"), lpaStore))
	handleLpa(page.Paths.WhatYoullNeedToConfirmYourIdentity, CanGoBack,
		page.Guidance(tmpls.Get("what_youll_need_to_confirm_your_identity.gohtml"), lpaStore))

	for path, page := range map[string]int{
		page.Paths.SelectYourIdentityOptions:  0,
		page.Paths.SelectYourIdentityOptions1: 1,
		page.Paths.SelectYourIdentityOptions2: 2,
	} {
		handleLpa(path, CanGoBack,
			SelectYourIdentityOptions(tmpls.Get("select_your_identity_options.gohtml"), lpaStore, page))
	}

	handleLpa(page.Paths.YourChosenIdentityOptions, CanGoBack,
		YourChosenIdentityOptions(tmpls.Get("your_chosen_identity_options.gohtml"), lpaStore))
	handleLpa(page.Paths.IdentityWithYoti, CanGoBack,
		IdentityWithYoti(tmpls.Get("identity_with_yoti.gohtml"), lpaStore, yotiClient, yotiScenarioID))
	handleLpa(page.Paths.IdentityWithYotiCallback, CanGoBack,
		IdentityWithYotiCallback(tmpls.Get("identity_with_yoti_callback.gohtml"), yotiClient, lpaStore))
	handleLpa(page.Paths.IdentityWithOneLogin, CanGoBack,
		IdentityWithOneLogin(logger, oneLoginClient, sessionStore, random.String))
	handleLpa(page.Paths.IdentityWithOneLoginCallback, CanGoBack,
		IdentityWithOneLoginCallback(tmpls.Get("identity_with_one_login_callback.gohtml"), oneLoginClient, sessionStore, lpaStore))

	for path, identityOption := range map[string]identity.Option{
		page.Paths.IdentityWithPassport:                 identity.Passport,
		page.Paths.IdentityWithBiometricResidencePermit: identity.BiometricResidencePermit,
		page.Paths.IdentityWithDrivingLicencePaper:      identity.DrivingLicencePaper,
		page.Paths.IdentityWithDrivingLicencePhotocard:  identity.DrivingLicencePhotocard,
		page.Paths.IdentityWithOnlineBankAccount:        identity.OnlineBankAccount,
	} {
		handleLpa(path, CanGoBack,
			IdentityWithTodo(tmpls.Get("identity_with_todo.gohtml"), identityOption))
	}

	handleLpa(page.Paths.ReadYourLpa, CanGoBack,
		page.Guidance(tmpls.Get("read_your_lpa.gohtml"), lpaStore))
	handleLpa(page.Paths.YourLegalRightsAndResponsibilities, CanGoBack,
		page.Guidance(tmpls.Get("your_legal_rights_and_responsibilities.gohtml"), lpaStore))
	handleLpa(page.Paths.SignYourLpa, CanGoBack,
		SignYourLpa(tmpls.Get("sign_your_lpa.gohtml"), lpaStore))
	handleLpa(page.Paths.WitnessingYourSignature, CanGoBack,
		WitnessingYourSignature(tmpls.Get("witnessing_your_signature.gohtml"), lpaStore, notifyClient, random.Code, time.Now))
	handleLpa(page.Paths.WitnessingAsCertificateProvider, CanGoBack,
		WitnessingAsCertificateProvider(tmpls.Get("witnessing_as_certificate_provider.gohtml"), lpaStore, shareCodeSender, time.Now))
	handleLpa(page.Paths.YouHaveSubmittedYourLpa, CanGoBack,
		page.Guidance(tmpls.Get("you_have_submitted_your_lpa.gohtml"), lpaStore))

	handleLpa(page.Paths.Progress, CanGoBack,
		page.Guidance(tmpls.Get("lpa_progress.gohtml"), lpaStore))
}

type handleOpt byte

const (
	None handleOpt = 1 << iota
	RequireSession
	CanGoBack
)

func makeHandle(mux *http.ServeMux, store sesh.Store, defaultOptions handleOpt, errorHandler page.ErrorHandler) func(string, handleOpt, page.Handler) {
	return func(path string, opt handleOpt, h page.Handler) {
		opt = opt | defaultOptions

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			appData := page.AppDataFromContext(ctx)
			appData.Page = path
			appData.CanGoBack = opt&CanGoBack != 0

			if opt&RequireSession != 0 {
				session, err := sesh.Donor(store, r)
				if err != nil {
					http.Redirect(w, r, page.Paths.Start, http.StatusFound)
					return
				}

				appData.SessionID = base64.StdEncoding.EncodeToString([]byte(session.Sub))

				data := page.SessionDataFromContext(ctx)
				if data != nil {
					data.SessionID = appData.SessionID
					ctx = page.ContextWithSessionData(ctx, data)

					appData.LpaID = data.LpaID
				} else {
					ctx = page.ContextWithSessionData(ctx, &page.SessionData{SessionID: appData.SessionID})
				}
			}

			if err := h(appData, w, r.WithContext(page.ContextWithAppData(ctx, appData))); err != nil {
				errorHandler(w, r, err)
			}
		})
	}
}

func routeToLpa(mux http.Handler, notFoundHandler page.Handler) http.HandlerFunc {
	const prefixLength = len("/lpa/")

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
		if len(r.URL.RawPath) > prefixLength+len(id) {
			r2.URL.RawPath = r.URL.RawPath[prefixLength+len(id):]
		}

		mux.ServeHTTP(w, r2.WithContext(page.ContextWithSessionData(r2.Context(), &page.SessionData{LpaID: id})))
	}
}
